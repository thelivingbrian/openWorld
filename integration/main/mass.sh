#!/bin/bash

# Loop through values of A and B
for A in {0..1}; do
  for B in {0..4}; do
    for TEAM in "team-blue" "team-fuchsia"; do
      # Construct the stagename
      STAGENAME="${TEAM}:${A}-${B}"
      # Execute the curl request
      curl -X POST "http://localhost:4440/mass?stagename=${STAGENAME}&read=true&count=25&ttl=3600&team=fuchsia"
    done
  done
done

