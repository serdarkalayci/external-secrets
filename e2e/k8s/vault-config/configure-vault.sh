#!/bin/sh
set -euxo pipefail;

export VAULT_TOKEN=${1}

# ------------------
#   SECRET BACKENDS
# ------------------
vault secrets enable -path=secret -version=2 kv
vault secrets enable -path=secret_v1 -version=1 kv

# ------------------
#   CERT AUTH
#   https://www.vaultproject.io/docs/auth/cert
# ------------------
vault auth enable cert
vault policy write \
    external-secrets-operator \
    /etc/vault-config/vault-policy-es.hcl

vault write auth/cert/certs/external-secrets-operator \
    display_name=external-secrets-operator \
    policies=external-secrets-operator \
    certificate=@/etc/vault-config/es-client.pem \
    ttl=3600

# test certificate login
unset VAULT_TOKEN
vault login \
    -client-cert=/etc/vault-config/es-client.pem \
    -client-key=/etc/vault-config/es-client-key.pem \
    -method=cert \
    name=external-secrets-operator

vault kv put secret/foo/bar baz=bang
vault kv get secret/foo/bar

# ------------------
#   App Role AUTH
#   https://www.vaultproject.io/docs/auth/approle
# ------------------
export VAULT_TOKEN=${1}
vault auth enable -path=myapprole approle

vault write auth/myapprole/role/eso-e2e-role \
    secret_id_ttl=10m \
    token_num_uses=10 \
    token_policies=external-secrets-operator \
    token_ttl=1h \
    token_max_ttl=4h \
    secret_id_num_uses=40
