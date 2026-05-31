<img width="122" height="131" alt="image" src="https://github.com/user-attachments/assets/9781143d-2eab-47d5-8adf-9598a7b239ad" />

# OIDC Broker
This tool acts as a broker that allows you to incorporate output from another OIDC provider into a newly generated token. This token can then be configured as a trusted entity for the Kubernetes API.

## Description
This tool works seamlessly alongside the Tocrocon Docker component, which is also part of the k8stooling suite.


### Configuration of deployment file
The following enviroment variables can be set as part of the Deployment
``````
- name: ISSUER_URL
  value: "{{OIDC_BROKER_URL}}"
- name: TENANT_ID
  value: "{{YOUR TENANT_ID}}"
- name: TOKEN_TTL
  value: "1h"
``````
- <b>TENANT_ID</b> is optional and has no meaning.
- <b>TOKEN_TTL</b> is default 900s in case nothing is set


### Configuration of Secrets file
The Secret Name in the Namespace needs to be <b>oidc-broker</b>
The Key which represents the Signing RSA has to be set to <b>rsa_key</b>

### Deployment YAML-Files
(!) <b><i>Check also the yaml-Examples files</b></i>