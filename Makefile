DATASETS = elliptic++ act-mooc roadNet-CA

.PHONY: clean

clean: $(DATASETS)

$(DATASETS):
	rm -f datasets/$@/converted.rdf
