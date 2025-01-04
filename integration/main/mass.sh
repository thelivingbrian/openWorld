#!/bin/bash

for A in {0..1}; do
  for B in {0..3}; do
    for TEAM in "team-blue" "team-fuchsia"; do
      # Construct the stagename
      STAGENAME="${TEAM}:${A}-${B}"
      # Execute the curl request
      curl -X POST "http://localhost:4440/mass?stagename=${STAGENAME}&read=true&count=16&ttl=3600&team=fuchsia"
    done
  done
done

for A in {2..3}; do
  for B in {4..7}; do
    for TEAM in "team-blue" "team-fuchsia"; do
      # Construct the stagename
      STAGENAME="${TEAM}:${A}-${B}"
      # Execute the curl request
      curl -X POST "http://localhost:4440/mass?stagename=${STAGENAME}&read=true&count=16&ttl=3600&team=sky-blue"
    done
  done
done

