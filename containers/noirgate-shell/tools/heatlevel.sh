#!/bin/zsh
ip=$(curl -s ip.me)
# risk=$(anubis-cli -ip "$ip" | jq -r .RiskLevel)
risk=High
if [ $risk = "Low" ]; then
    printf "[$ip โ ]"
elif [ $risk = "Medium" ]; then
    printf "[$ip โ ๏ธ ]"
elif [ $risk = "High" ]; then
    printf "[$ip ๐ฅ ]"
else
    printf "[$ip ๐คจโ ]"
fi