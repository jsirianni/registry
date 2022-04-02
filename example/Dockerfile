FROM hashicorp/terraform:1.1.7

COPY scripts/tls/test-ca.crt /usr/local/share/ca-certificates/test-ca.crt
COPY main.tf .
RUN update-ca-certificates
