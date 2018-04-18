
FROM tianon/ubuntu-core:14.04
ENV ONTOLOGY_PATH /var/ontology
RUN mkdir -p $ONTOLOGY_PATH
COPY ontology $ONTOLOGY_PATH
ADD config.tgz $ONTOLOGY_PATH
WORKDIR $ONTOLOGY_PATH
CMD ["./ontology","-p","passwordtest"]

