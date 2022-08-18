#!/bin/zsh
ip=$(curl -s ip.me)
# risk=$(anubis-cli -ip "$ip" | jq -r .RiskLevel)
risk=High
if [ $risk = "Low" ]; then
    printf "[$ip ✅ ]"
elif [ $risk = "Medium" ]; then
    printf "[$ip ⚠️ ]"
elif [ $risk = "High" ]; then
    printf "[$ip 🔥 ]"
else
    printf "[$ip 🤨❓ ]"
fi