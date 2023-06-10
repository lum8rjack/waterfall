FROM ubuntu:22.04

RUN apt-get update -y

# Install dependencies (nmap is not the latest - could install snap and then nmap)
RUN apt-get install -y git wget libpcap-dev nmap chromium-browser libnss3 libatk1.0-dev libatk-bridge2.0-dev libcups2-dev libxcomposite-dev libxdamage-dev libxrandr-dev libgbm-dev libxkbcommon-dev libpango1.0-dev libasound2-dev

# Install Go
RUN wget https://go.dev/dl/go1.20.4.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.20.4.linux-amd64.tar.gz
RUN rm go1.20.4.linux-amd64.tar.gz

# Setup environment variables
ENV GOROOT=/usr/local/go
ENV GOPATH=/root/go
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Install tools
RUN go install github.com/projectdiscovery/httpx/cmd/httpx@latest
RUN go install github.com/projectdiscovery/katana/cmd/katana@latest
RUN go install github.com/projectdiscovery/notify/cmd/notify@latest
RUN go install github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest
RUN go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest
RUN go install github.com/projectdiscovery/tlsx/cmd/tlsx@latest
RUN go install github.com/sensepost/gowitness@latest
RUN go install github.com/projectdiscovery/naabu/v2/cmd/naabu@latest
RUN go install github.com/projectdiscovery/dnsx/cmd/dnsx@latest

# Test run nuclei to install templates and headless
RUN nuclei -tl -tags homeassistant -headless

# Set working directory
WORKDIR /opt/waterfall

# Copy files
COPY . /opt/waterfall/

# Build
RUN go build -o waterfall.bin .

ENTRYPOINT ["/opt/waterfall/waterfall.bin", "-config", "config.yaml"]
