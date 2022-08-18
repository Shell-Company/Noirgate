#  add the rule below to your container host to prevent SSRF against meta-data 
#  iptables -A DOCKER-USER -d 169.254.169.254 -p tcp -j DROP
FROM golang:1.18
ENV DEBIAN_FRONTEND=noninteractive
RUN wget -q https://packages.microsoft.com/config/ubuntu/16.04/packages-microsoft-prod.deb
RUN dpkg -i packages-microsoft-prod.deb
RUN apt update && apt upgrade -y && apt install tor ssh -y && apt install git jq nano nmap lynx w3m curl python python3-pip iptables mitmproxy zsh awscli net-tools vim netcat ncat whois hexedit zip htop npm qrencode forensics-all powershell privoxy build-essential patch ruby ruby-dev libsqlite3-dev libpcap-dev libpq-dev zlib1g-dev libz-dev liblzma-dev tmux gdb neovim -y
WORKDIR /usr/bin
# Install kubectl
RUN curl -LO https://dl.k8s.io/release/v1.21.0/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl
# install npm and graphqurl
RUN npm install -g graphqurl
# # install mssql-cli-
RUN pip3 install mssql-cli pymssql oletools
# install pgcli
RUN pip3 install pgcli --no-deps
# install azure cli tools 
RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

WORKDIR /opt 
# install metasploit framework
RUN gem install bundler -v 2.2.27
RUN git clone https://github.com/rapid7/metasploit-framework.git && cd  ./metasploit-framework && bundle update --bundler && bundle install

# clean up packages
RUN apt clean
# configure user sandbox
RUN useradd -m anon -s /bin/zsh
RUN git clone --depth=1 https://github.com/cqsd/gotty && cd gotty && GO11MODULE=on go install . && cd - && rm -fr /tmp/gotty /tmp/hsperfdata_root
WORKDIR /tmp
# install radare2
WORKDIR /opt
RUN git clone https://github.com/radareorg/radare2 && cd radare2 && ./sys/install.sh && rm -fr /tmp/radare2

USER root
ENV TERM xterm 
ENV TORSOCKS_LOG_LEVEL 1
EXPOSE 8080/tcp
# Enable DNS over TLS 
ADD ./resolved.conf /etc/systemd/resolved.conf
ADD ./dfir.txt /etc/motd
# ADD ./welcome /etc/motd
ADD ./tools/aterm /usr/bin/aterm
ADD ./tools/pdfid.py /usr/bin/pdfid.py
ADD ./tools/anubis-cli /usr/bin/anubis
ADD ./limits.conf /etc/security/limits.conf
ADD ./tools/heatlevel.sh /usr/bin/heatlevel.sh
ADD ./tools/readme /opt/readme
# Build noirgate-otp and install it in /usr/bin
WORKDIR /build
ADD ./noirgate-otp .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -extldflags "-static"' -o /usr/bin/noirgate-otp . && rm -rf /build
WORKDIR /tmp
# update zshrc
USER anon
RUN  sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
# Fix permissions for zshrc
USER root
ADD ./zshrc /home/anon/.zshrc 
RUN chown anon:anon /home/anon/.zshrc
# install eksctl
RUN curl --silent --location "https://github.com/weaveworks/eksctl/releases/latest/download/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp &&  mv /tmp/eksctl /usr/local/bin

# install google cloud cli 
USER anon
WORKDIR /home/anon
RUN curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-356.0.0-linux-x86_64.tar.gz && tar -xf ./google-cloud-sdk-356.0.0-linux-x86_64.tar.gz && bash ./google-cloud-sdk/install.sh --quiet --rc-path "/home/anon/.zshrc" && rm ./google-cloud-sdk-356.0.0-linux-x86_64.tar.gz
USER root
# install k9s
RUN git clone https://github.com/derailed/k9s && cd k9s && go mod tidy && go build && mv ./k9s /usr/bin/k9s  && rm -rf /tmp/k9s
#install privoxy
ADD ./privoxy.conf /etc/privoxy/config
ENV HTTP_PROXY=http://127.0.0.1:8118
# no mans land (noirgate-shell)
RUN echo "           built on $(date) by Shell Company" >> /etc/motd
WORKDIR /tmp
# ENTRYPOINT service tor start && service privoxy start &&  su anon -c "LD_PRELOAD=/usr/lib/x86_64-linux-gnu/torsocks/libtorsocks.so /go/bin/gotty --title-format noirgate🌀 -w bash -c 'noirgate-otp && zsh'"
ENTRYPOINT service tor start && service privoxy start &&  su anon -c "/go/bin/gotty --title-format noirgate🌀 -w bash -c 'noirgate-otp && zsh'"