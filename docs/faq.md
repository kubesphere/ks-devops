Please feel free to checkout the following frequently asked questions, hopefully this document could help you.

## Pipeline cannot be triggered
Effected versions: `v3.2.1`

| Possible reason | Potential solution |
|---|---|
| The account was not synced due to the LDAP issues | Reset your account's password |
| The token of Jenkins is incorrect | Restart the deployment `devops-controller` and `devops-apiserver` if you didn't change the token manually |
