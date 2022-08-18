#!/bin/bash

#prompt user for subdomain
# echo "Enter subdomain:"
# read subdomain

# #prompt user for top level domain
# echo "Enter top level domain:"
# read tld
# copy template nginx conf file to nginx conf directory
cp ./nginx/nginx.conf.template ./nginx/nginx.conf

#replace subdomain and tld in template file with user supplied values
sed -i "s/NOIRGATE_SUBDOMAIN/$NOIRGATE_SUB/g" ./nginx/nginx.conf 
sed -i "s/NOIRGATE_TLD/$NOIRGATE_TLD/g" ./nginx/nginx.conf 

#update static web app 
cp ./nginx/index.html.template ./nginx/index.html
sed -i "s/NOIRGATE_SUBDOMAIN/$NOIRGATE_SUB/g" ./nginx/html/index.html
sed -i "s/NOIRGATE_TLD/$NOIRGATE_TLD/g" ./nginx/html/index.html
