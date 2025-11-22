# powershell -ExecutionPolicy Bypass -File .\mass.ps1

# First loop: A = 0..1, B = 0..3
foreach ($A in 0..1) {
    foreach ($B in 0..3) {
        foreach ($TEAM in "team-blue", "team-fuchsia") {

            $STAGENAME = "${TEAM}:${A}-${B}"

            curl.exe -X POST "http://localhost:4440/mass?stagename=$STAGENAME&read=true&count=16&ttl=1800&action=random&team=fuchsia"
        }
    }
}

# Second loop: A = 2..3, B = 4..7
foreach ($A in 2..3) {
    foreach ($B in 4..7) {
        foreach ($TEAM in "team-blue", "team-fuchsia") {

            $STAGENAME = "${TEAM}:${A}-${B}"

            curl.exe -X POST "http://localhost:4440/mass?stagename=$STAGENAME&read=true&count=16&ttl=1800&action=random&team=sky-blue"
        }
    }
}

