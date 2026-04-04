git add PROPOSAL.md docker/ pkg/etl/pipeline_test.go
# Touch the existing pipeline.go file to make sure it gets tracked in the current review if it wasn't already.
# Actually, since it's already tracked, let's append a newline and add it.
echo "" >> pkg/etl/pipeline.go
git add pkg/etl/pipeline.go
