#  add the rule below to your container host to prevent SSRF against meta-data 
#  iptables -A DOCKER-USER -d 169.254.169.254 -p tcp -j DROP
FROM golang:1.21.1
ENV DEBIAN_FRONTEND=noninteractive
RUN wget -q https://packages.microsoft.com/config/ubuntu/16.04/packages-microsoft-prod.deb
RUN dpkg -i packages-microsoft-prod.deb
RUN apt update && apt upgrade -y && apt install tor ssh -y && apt install git jq nano nasm nmap lynx w3m curl python3 python3-pip iptables mitmproxy zsh awscli net-tools vim ncat whois hexedit zip htop npm qrencode forensics-all powershell privoxy build-essential patch ruby ruby-dev libsqlite3-dev libpcap-dev ruby3.1 ruby3.1-dev libpq-dev zlib1g-dev libz-dev liblzma-dev tmux gdb neovim skopeo  -y --no-install-recommends
WORKDIR /usr/bin
# RUN pip3 install mssql-cli pymssql oletools  --upgrade
# Install kubectl
RUN curl -LO https://dl.k8s.io/release/v1.21.0/bin/linux/amd64/kubectl && chmod +x /usr/bin/kubectl
# install npm and graphqurl
RUN npm install -g graphqurl 
# install pgcli
# RUN pip3 install pgcli  
# install azure cli tools 
# RUN curl -sL https://aka.ms/InstallAzureCLIDeb | bash

WORKDIR /opt 
# install metasploit framework
RUN gem install bundler -v 2.2.27
RUN gem install xmlrpc
RUN git clone --depth 1 https://github.com/rapid7/metasploit-framework.git 
RUN cd  ./metasploit-framework && bundle update --bundler && bundle install

# clean up packages
RUN apt clean
# configure user sandbox
RUN useradd -m anon -s /bin/zsh
RUN git clone --depth=1 https://github.com/cqsd/gotty && cd gotty && GO11MODULE=on go install . && cd - && rm -fr /tmp/gotty /tmp/hsperfdata_root
WORKDIR /tmp
# install radare2
WORKDIR /opt
RUN git clone --depth 1 https://github.com/radareorg/radare2 && cd radare2 && ./sys/install.sh && rm -fr /tmp/radare2

USER root
ENV TERM xterm 
ENV TORSOCKS_LOG_LEVEL 1
EXPOSE 8080/tcp
# Enable DNS over TLS 
ADD ./resolved.conf /etc/systemd/resolved.conf
ADD ./noirgate.txt /etc/motd
# ADD ./welcome /etc/motd
ADD ./tools/aterm /usr/bin/aterm
ADD ./tools/pdfid.py /usr/bin/pdfid.py
ADD ./limits.conf /etc/security/limits.conf
ADD ./tools/heatlevel.sh /usr/bin/heatlevel.sh
ADD ./tools/readme /opt/readme
ADD ./tools/ophanim /usr/bin/ophanim 
RUN  chmod +x /usr/bin/ophanim

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
RUN curl -O https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-436.0.0-linux-x86_64.tar.gz && tar -xf ./google-cloud-sdk-436.0.0-linux-x86_64.tar.gz && bash ./google-cloud-sdk/install.sh --quiet --rc-path "/home/anon/.zshrc" && rm ./google-cloud-sdk-436.0.0-linux-x86_64.tar.gz
# install uncover
ENV GOPATH=/home/anon/go
# projectdiscovery tools
RUN go install -v github.com/projectdiscovery/uncover/cmd/uncover@latest
RUN go install -v github.com/projectdiscovery/nuclei/cmd/nuclei@latest
RUN go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest
RUN go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
RUN	git config --global --add safe.directory /opt/metasploit-framework
RUN pip3 install -U pacu --break-system-packages
RUN go install -v github.com/owasp-amass/amass/v3/...@master
RUN go install -v github.com/datadog/stratus-red-team/v2/cmd/stratus@latest
USER root
# install k9s
RUN git clone --depth 1 https://github.com/derailed/k9s && cd k9s && go mod tidy && go build && mv ./k9s /usr/bin/k9s  && rm -rf /tmp/k9s
#install privoxy
ADD ./privoxy.conf /etc/privoxy/config
ENV HTTP_PROXY=http://127.0.0.1:8118
# no man's lan (noirgate-shell)
RUN git clone --depth 1 https://gitlab.com/exploit-database/exploitdb.git /opt/exploit-database
RUN ln -sf /opt/exploit-database/searchsploit /usr/local/bin/searchsploit
RUN mkdir /run/containers
RUN chmod 777 /run/containers
RUN chown -R anon:anon /home/anon
RUN apt clean
WORKDIR /tmp
RUN rm .cc.txt
USER anon
RUN mkdir /home/anon/.config/ophanim/ && touch /home/anon/.config/ophanim/.prompt_history
# install iamlive
RUN go install github.com/iann0036/iamlive@latest
USER root
# install ophanim research 
ADD ./tools/research.sh /usr/bin/research
RUN chmod +x /usr/bin/research
# install cloudgrep
RUN wget https://github.com/cado-security/cloudgrep/releases/download/v105/dist-linux.zip && unzip ./dist-linux.zip && chmod +x ./cloudgrep && mv ./cloudgrep /usr/local/bin/cloudgrep && rm -rf dist-linux.zip
# install cloudflow
RUN wget https://github.com/louisbarrett/cloudflow/releases/download/v0.1.1/cloudflow_Linux_x86_64.tar.gz && tar -xf ./cloudflow_Linux_x86_64.tar.gz && chmod +x ./cloudflow && mv ./cloudflow /usr/local/bin/cloudflow && rm -rf cloudflow_Linux_x86_64.tar.gz
RUN echo "           built on $(date) by AZAI Security Research" >> /etc/motd
ENTRYPOINT service tor start && service privoxy start &&  su anon -c "/go/bin/gotty --title-format 'AZAI Ophanim' -w bash -c 'noirgate-otp && torsocks --shell'"
