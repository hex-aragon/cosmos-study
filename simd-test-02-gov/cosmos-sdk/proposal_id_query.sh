
#export PROPOSAL_ID=$(simd query tx 7A52F6FB0BEEB0EC6A48F08DDBDC17E4557170BF359D8ECA6F158DAAF6DA9C0B --output json | jq '.events' | jq -r '.[] | select(.type == "submit_proposal") | .attributes[0].value' | jq -r '.')


export PROPOSAL_ID=$(simd query tx DF658DB937EC1C541A273BF9CC2BCDE10E8A98C64374C0632E3BF183DA0E0D6F --output json | jq '.events' | jq -r '.[] | select(.type == "submit_proposal") | .attributes[0].value' | jq -r '.')