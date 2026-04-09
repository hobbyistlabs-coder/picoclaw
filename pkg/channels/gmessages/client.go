package gmessages

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/rs/zerolog"
	"go.mau.fi/mautrix-gmessages/pkg/libgm"
	"go.mau.fi/mautrix-gmessages/pkg/libgm/events"
	"go.mau.fi/mautrix-gmessages/pkg/libgm/gmproto"

	"jane/pkg/logger"
	"jane/pkg/runtimepaths"
)

type SessionData struct {
	AuthDataJSON json.RawMessage `json:"auth_data"`
	PushKeysJSON json.RawMessage `json:"push_keys,omitempty"`
}

type GMClient struct {
	GM          *libgm.Client
	SessionPath string
}

func (c *GMessagesChannel) initClient(ctx context.Context) error {
	dataDir := c.cfg.DataDir
	if dataDir == "" {
		dataDir = filepath.Join(runtimepaths.HomeDir(), "gmessages")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create gmessages data dir: %w", err)
	}

	sessionPath := filepath.Join(dataDir, "session.json")
	dbPath := filepath.Join(dataDir, "messages.db")

	// Initialize DB (we will fill this in db.go)
	store, err := NewStore(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize db: %w", err)
	}

	c.state.store = store

	// Init LibGM Logger using zerolog (placeholder since GetLogger isn't exported in jane/pkg/logger)
	// For production we can proxy mautrix logs to our system logger but keeping it simple for now.
	gmLogger := zerolog.Nop()

	var client *GMClient

	// Try loading session
	sessionData, err := loadSession(sessionPath)
	if err == nil {
		logger.InfoCF("channels.gmessages", "Loaded existing session", nil)
		authData := libgm.NewAuthData()
		if err := json.Unmarshal(sessionData.AuthDataJSON, authData); err != nil {
			return fmt.Errorf("unmarshal auth data: %w", err)
		}

		var pushKeys *libgm.PushKeys
		if len(sessionData.PushKeysJSON) > 0 {
			pushKeys = &libgm.PushKeys{}
			if err := json.Unmarshal(sessionData.PushKeysJSON, pushKeys); err != nil {
				return fmt.Errorf("unmarshal push keys: %w", err)
			}
		}

		gmClient := libgm.NewClient(authData, pushKeys, gmLogger)
		client = &GMClient{GM: gmClient, SessionPath: sessionPath}
	} else {
		logger.InfoCF("channels.gmessages", "No session found or failed to load. Initiating pairing", map[string]any{"err": err.Error()})
		authData := libgm.NewAuthData()
		gmClient := libgm.NewClient(authData, nil, gmLogger)
		client = &GMClient{GM: gmClient, SessionPath: sessionPath}

		if err := c.handlePairing(ctx, client); err != nil {
			return err
		}
	}

	c.state.client = client

	// Setup event handler
	handler := &EventHandler{
		Store:       store,
		Channel:     c,
		SessionPath: sessionPath,
		Client:      client,
	}

	client.GM.SetEventHandler(handler.Handle)

	// Connect
	if err := client.GM.Connect(); err != nil {
		return fmt.Errorf("failed to connect libgm: %w", err)
	}

	logger.InfoCF("channels.gmessages", "gmessages client connected successfully", nil)

	return nil
}

func (c *GMessagesChannel) stopClient(ctx context.Context) {
	if c.state.client != nil {
		client := c.state.client
		if client.GM != nil {
			client.GM.Disconnect()
		}
	}
	if c.state.store != nil {
		store := c.state.store
		store.Close()
	}
}

func (c *GMClient) SessionData() (*SessionData, error) {
	authJSON, err := json.Marshal(c.GM.AuthData)
	if err != nil {
		return nil, fmt.Errorf("marshal auth data: %w", err)
	}
	var pushJSON json.RawMessage
	if c.GM.PushKeys != nil {
		pushJSON, err = json.Marshal(c.GM.PushKeys)
		if err != nil {
			return nil, fmt.Errorf("marshal push keys: %w", err)
		}
	}
	return &SessionData{
		AuthDataJSON: authJSON,
		PushKeysJSON: pushJSON,
	}, nil
}

func loadSession(path string) (*SessionData, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s SessionData
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func saveSession(path string, data *SessionData) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

func (c *GMessagesChannel) handlePairing(ctx context.Context, client *GMClient) error {
	pairingCh := make(chan struct{})
	var pairErr error

	handler := func(rawEvt any) {
		switch evt := rawEvt.(type) {
		case *events.PairSuccessful:
			logger.InfoCF("channels.gmessages", "Pairing successful", map[string]any{"phone_id": evt.PhoneID})

			// Save session
			sessionData, err := client.SessionData()
			if err != nil {
				logger.ErrorCF("channels.gmessages", "Failed to get session data", map[string]any{"error": err.Error()})
			} else {
				if err := saveSession(client.SessionPath, sessionData); err != nil {
					logger.ErrorCF("channels.gmessages", "Failed to save session", map[string]any{"error": err.Error()})
				} else {
					logger.InfoCF("channels.gmessages", "Session saved to file", map[string]any{"path": client.SessionPath})
				}
			}

			close(pairingCh)
		case *events.ListenFatalError:
			pairErr = evt.Error
			close(pairingCh)
		}
	}

	client.GM.SetEventHandler(handler)

	logger.InfoCF("channels.gmessages", "Starting pairing process...", nil)

	var pairErr2 error
	pairCB := func(data *gmproto.PairedData) {
		logger.InfoCF("channels.gmessages", "Pairing successful", map[string]any{"phone_id": data.GetMobile().GetSourceID()})

		sessionData, err := client.SessionData()
		if err != nil {
			logger.ErrorCF("channels.gmessages", "Failed to get session data", map[string]any{"error": err.Error()})
		} else {
			if err := saveSession(client.SessionPath, sessionData); err != nil {
				logger.ErrorCF("channels.gmessages", "Failed to save session", map[string]any{"error": err.Error()})
			} else {
				logger.InfoCF("channels.gmessages", "Session saved to file", map[string]any{"path": client.SessionPath})
			}
		}

		close(pairingCh)
	}
	client.GM.PairCallback.Store(&pairCB)

	qrURL, err := client.GM.StartLogin()
	if err != nil {
		return fmt.Errorf("failed to start pairing login: %w", err)
	}

	fmt.Println("\n=== Scan this QR code in Google Messages (Device Pairing) ===")
	qrterminal.GenerateHalfBlock(qrURL, qrterminal.L, os.Stdout)
	fmt.Println("Waiting for pairing...")

	select {
	case <-pairingCh:
		if pairErr != nil {
			return fmt.Errorf("pairing failed: %w", pairErr)
		}
		if pairErr2 != nil {
			return fmt.Errorf("pairing failed: %w", pairErr2)
		}
		// pairing successful, we disconnect so the main initClient can set the real event handler and reconnect
		client.GM.Disconnect()
		return nil
	case <-ctx.Done():
		client.GM.Disconnect()
		return ctx.Err()
	case <-time.After(5 * time.Minute):
		client.GM.Disconnect()
		return fmt.Errorf("pairing timed out")
	}
}
