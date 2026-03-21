## v0.0.3 - 2026-03-21

* Fix cookie secret to be 32-bytes base64 encoded , add KEYCLOAK_HOST to configmap (#28)
* Move ClientID to the generated Secret instead of the ConfigMap (#30)

## v0.0.2 - 2026-03-19

* Improve image release and README fixes (#23)
* Make envVar format Secret and ConfigMap keys optional (#24)
* Use Helm Job to signal webhooks are ready, also support deploying imagePullSecret (#27)

## v0.0.1 - 2026-03-18

* Initial release
