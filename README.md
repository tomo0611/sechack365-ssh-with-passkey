# ssh-with-passkey (SecHack365)

## フォルダ構成 (Folder Structure)

- pam-module: PAM Module のソースコード / Source code of PAM Module
- auth-server: 認証サーバのソースコード / Source code of Authentication Server

## ビルド方法 (How to build PAM Module)

```bash
cd pam-module
sudo apt install plocate
locate pam_unix.so
# install cmake, make, and other required packages
sudo apt install cmake make libpam0g-dev
# check your pam_unix.so file location
# /usr/lib/x86_64-linux-gnu/security/pam_unix.so
cd build

# If you use cmake
cmake ..

# Not use cmake
gcc -fPIC -fno-stack-protector -c src/main.c
sudo ld -x --shared -o /lib/x86_64-linux-gnu/security/pam_alya.so main.o
```

## インストール方法 (How to install PAM Module)

```bash
# soファイルの場所に合わせて要調整 / You need to adjust the path of so file
sudo cp main.so /usr/lib/x86_64-linux-gnu/security/pam_alya.so
# QRコード生成用のライブラリをインストール / Install a library for generating QR code
sudo apt install go-qrcode
```

`/etc/pam.d/sshd`の一番上に以下の行を追加 (Add this line end of `/etc/pam.d/sshd`):

```
# For Passkey Based Login
auth    [success=done user_unknown=ignore]      pam_alya.so
```

## ビルド方法 (How to build Authentication Server)

```bash
cd auth-server
go build
```

## インストール方法 (How to install Authentication Server)

```bash
# /home/tomo/sechack365-ssh-with-passkey/auth-server/sechack365-ssh-with-passkey を実行できるように設定 (実行ユーザーはtomoでよい) / Set sechack365-ssh-with-passkey to be executable
# WorkingDirectoryは /home/tomo/sechack365-ssh-with-passkey/auth-server に設定するのわすれないように！(webディレクトリが見れなくなる) / Set WorkingDirectory to /home/tomo/sechack365-ssh-with-passkey/auth-server (Don't forget!)
sudo nano /etc/systemd/system/passkey-auth-server.service
sudo systemctl start passkey-auth-server
sudo systemctl enable passkey-auth-server
sudo adduser alya
# sshd_configでKbdInteractiveAuthentication yesに変更 (キーボードからの認証を有効にする) / Change KbdInteractiveAuthentication to yes in sshd_config (Enable keyboard-interactive authentication)
# UsePAM yesに変更 (PAMを使う) / Change UsePAM to yes (Use PAM)
sudo nano /etc/ssh/sshd_config
# sudo systemctl restart ssh じゃだめだったので再起動 / sudo systemctl restart ssh didn't work, so I restarted
sudo reboot now
sudo apt install nginx
# proxy_pass http://127.0.0.1:1323/; と proxy_read,connect,send_timeoutを90に設定 / And Set timeout to 90
# proxy_set_header X-Forwarded-for $remote_addr; を追加 / Add proxy_set_header X-Forwarded-for $remote_addr;
sudo nano /etc/nginx/sites-available/passkey.tomo0611.dev
sudo ln -s /etc/nginx/sites-available/passkey.tomo0611.dev /etc/nginx/sites-enabled/passkey.tomo0611.dev
sudo systemctl restart nginx
# certbotでLet's Encryptから証明書を取得 / Get a certificate from Let's Encrypt with certbot
sudo certbot --nginx -d passkey.tomo0611.dev

# Install Go (ref: https://go.dev/wiki/Ubuntu )
sudo add-apt-repository ppa:longsleep/golang-backports
sudo apt update
sudo apt install golang-go
scp -r -p 20022 ./ tomo@passkey.tomo0611.dev:/home/tomo/ssh-login-with-passkey/
```

## 参考にさせていただいたコード等 (References)

以下のコードや記事を参考にさせていただきました。ありがとうございます！  
I appreciate the following code and articles. Thank you!

- [pam-tutorials](https://github.com/fedetask/pam-tutorials/tree/master)
- [PassKey in Go](https://dev.to/egregors/passkey-in-go-1efk)
