DATASETS = elliptic++ act-mooc

.PHONY: clean

clean: $(DATASETS)

$(DATASETS):
	rm -f transform-$@ datasets/$@/transformed/$@.rdf
