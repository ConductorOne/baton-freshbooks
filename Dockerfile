FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-freshbooks"]
COPY baton-freshbooks /