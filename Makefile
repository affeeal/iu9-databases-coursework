DATASETS := elliptic++ act-mooc roadNet-CA ERC20-stablecoins
MINIMAL_DATASET := internal/converter/testdata/minimal

.PHONY: check clean compose-config demo down fmt-check load-demo test-race up vet wait

check: fmt-check vet test-race compose-config

fmt-check:
	@test -z "$$(gofmt -l $$(find . -name '*.go' -type f))" || \
		{ echo "Run gofmt on the listed files:"; gofmt -l $$(find . -name '*.go' -type f); exit 1; }

vet:
	go vet ./...

test-race:
	go test -race ./...

compose-config:
	docker compose config --quiet

demo:
	go run ./cmd/converter -dataset-path $(MINIMAL_DATASET)
	@diff -u $(MINIMAL_DATASET)/expected.rdf $(MINIMAL_DATASET)/output.rdf

up:
	docker compose up --detach

wait:
	scripts/wait-for-dgraph.sh

load-demo: demo wait
	scripts/load-dataset.sh $(MINIMAL_DATASET)

down:
	docker compose down

clean:
	rm -f $(addprefix datasets/,$(addsuffix /output.rdf,$(DATASETS))) \
		$(MINIMAL_DATASET)/output.rdf
