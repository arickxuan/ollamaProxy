
#!/bin/bash
sudo mkdir -p /opt/soft/cliproxy

cd /opt/soft/cliproxy
# wget http://soft.s3.tebi.io/tmp/cliproxy
sudo curl -O http://soft.s3.tebi.io/tmp/cliproxy
chmod +x /opt/soft/cliproxy/cliproxy

cd /opt/soft/

curl -O  http://soft.s3.tebi.io/tmp/frp_0.62.1_darwin_arm64.tar.gz

sudo tar zxf frp_0.62.1_darwin_arm64.tar.gz

sudo mv frp_0.62.1_darwin_arm64 frp

sudo rm frp/frpc.toml

cd frp

sudo curl -s -O http://soft.s3.tebi.io/tmp/zfrpc.toml

cd /Library/LaunchDaemons/

sudo curl -O http://soft.s3.tebi.io/tmp/frp.plist

sudo launchctl enable system/frp
sudo launchctl load -w /Library/LaunchDaemons/frp.plist

cd /opt/soft/cliproxy

./cliproxy

