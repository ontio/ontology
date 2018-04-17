
FROM ubuntu-core
ENV ONTOLOGY_PATH /var/ontology
RUN mkdir -p $ONTOLOGY_PATH
COPY ontology $ONTOLOGY_PATH
COPY nodectl $ONTOLOGY_PATH
ADD config.tgz $ONTOLOGY_PATH
WORKDIR $ONTOLOGY_PATH
CMD ["./ontology","-p","passwordtest"]

