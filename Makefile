DATASETS = elliptic++ act-mooc roadNet-CA ERC20-stablecoins

.PHONY: clean

clean: $(DATASETS)

$(DATASETS):
	rm -f datasets/$@/output.rdf
