#!/bin/bash

set -xeu

# TODO: Add Install instructions act via https://github.com/nektos/act

# SECRETS:
# CARVEL_RELEASE_SCRIPTS_PAT / Push access to vmware-tanzu/carvel-release-scripts

# https://docs.github.com/en/developers/webhooks-and-events/webhooks/webhook-events-and-payloads
act push -e <(cat <<EOF
{
  "push": {
      "ref": "refs/tags/v0.0.0"
  }
}
EOF
) --job goreleaser --job carvel-release-scripts --env SKIP_PUBLISH='--skip-publish' --dryrun