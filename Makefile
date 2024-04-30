DATASETS = elliptic++ act-mooc

.PHONY: clean

clean: $(DATASETS)

$(DATASETS):
	rm transform-$@ datasets/$@/transformed/$@.rdf
