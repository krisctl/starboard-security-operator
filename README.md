# starboard-security-operator

[![GitHub Build Actions][build-action-img]][actions]
[![Coverage Status][cov-img]][cov]

## Getting Started

1. Define custom security resources used by Starboard:
   ```
   $ kubectl apply -f https://raw.githubusercontent.com/aquasecurity/starboard/master/kube/crd/vulnerabilities-crd.yaml \
     -f https://raw.githubusercontent.com/aquasecurity/starboard/master/kube/crd/configauditreports-crd.yaml \
     -f https://raw.githubusercontent.com/aquasecurity/starboard/master/kube/crd/ciskubebenchreports-crd.yaml \
     -f https://raw.githubusercontent.com/aquasecurity/starboard/master/kube/crd/kubehunterreports-crd.yaml
   ```
2. Create a Secret that holds configuration of the Aqua CSP scanner:
   ```
   $ kubectl create secret generic starboard-scanner-aqua \
     --namespace starboard \
     --from-literal OPERATOR_SCANNER_AQUA_CSP_USER=$AQUA_CONSOLE_USERNAME \
     --from-literal OPERATOR_SCANNER_AQUA_CSP_PASSWORD=$AQUA_CONSOLE_PASSWORD \
     --from-literal OPERATOR_SCANNER_AQUA_CSP_VERSION=$AQUA_VERSION \
     --from-literal OPERATOR_SCANNER_AQUA_CSP_HOST=http://csp-console-svc.aqua:8080
   ```
3. Create a Service Account used to run Aqua CSP scan Jobs:
   ```
   $ kubectl apply -f kube/starboard-scanner-aqua.yaml
   ```
4. Create a Deployment for the Starboard Security Operator:
   ```
   $ kubectl apply -f kube/starboard-security-operator.yaml
   ```

## Configuration

| Name                                    | Default              | Description |
|-----------------------------------------|----------------------|-------------|
| `OPERATOR_STARBOARD_NAMESPACE`          | `starboard`          | The default namespace for Starboard |
| `OPERATOR_NAMESPACE`                    | `default`            | The namespace watched by the operator |
| `OPERATOR_SCANNER_TRIVY_ENABLED`        | `true`               | The flag to enable Trivy vulnerability scanner |
| `OPERATOR_SCANNER_TRIVY_VERSION`        | `0.9.1`              | The version of Trivy to be used |
| `OPERATOR_SCANNER_AQUA_CSP_ENABLED`     | `false`              | The flag to enable Aqua CSP vulnerability scanner |
| `OPERATOR_SCANNER_AQUA_CSP_VERSION`     | `4.6`                | The version of Aqua CSP scannercli container image to be used |

[build-action-img]: https://github.com/aquasecurity/starboard-security-operator/workflows/build/badge.svg
[actions]: https://github.com/aquasecurity/starboard-security-operator/actions
[cov-img]: https://codecov.io/github/aquasecurity/starboard-security-operator/branch/master/graph/badge.svg
[cov]: https://codecov.io/github/aquasecurity/starboard-security-operator
