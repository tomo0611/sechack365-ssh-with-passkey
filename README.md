# ssh-with-passkey (SecHack365)

## フォルダ構成

- pam-module: PAM Module のソースコード
- auth-server: 認証サーバのソースコード

## ビルド方法 (PAM Module)

```bash
cd pam-module
sudo apt install plocate
locate pam_unix.so
# cmakeとかmakeとかのインストール
sudo apt install cmake make libpam0g-dev
# soファイルの場所を確認
# /usr/lib/x86_64-linux-gnu/security/pam_unix.so
cd build

# cmakeを使うなら
cmake ..

# 使わないなら
gcc -fPIC -fno-stack-protector -c src/main.c
sudo ld -x --shared -o /lib/x86_64-linux-gnu/security/pam_alya.so main.o
```

## インストール方法 (PAM Module)

```bash
# soファイルの場所に合わせて要調整
sudo cp main.so /usr/lib/x86_64-linux-gnu/security/pam_alya.so
# QRコード生成用のライブラリをインストール
sudo apt install go-qrcode
```

`/etc/pam.d/sshd`の一番上に以下の行を追加:

```
# For Passkey Based Login
auth    [success=done user_unknown=ignore]      pam_alya.so
```

## ビルド方法 (Authentication Server)

```bash
cd auth-server
go build
```

## インストール方法 (Authentication Server)

```bash
# /home/tomo/sechack365-ssh-with-passkey/auth-server/sechack365-ssh-with-passkey を実行できるように設定 (実行ユーザーはtomoでよい)
# WorkingDirectoryは /home/tomo/sechack365-ssh-with-passkey/auth-server に設定するのわすれないように！(webディレクトリが見れなくなる)
sudo nano /etc/systemd/system/passkey-auth-server.service
sudo systemctl start passkey-auth-server
sudo systemctl enable passkey-auth-server
sudo adduser alya
# sshd_configでKbdInteractiveAuthentication yesに変更 (キーボードからの認証を有効にする)
# UsePAM yesに変更 (PAMを使う)
sudo nano /etc/ssh/sshd_config
# sudo systemctl restart ssh じゃだめだったので再起動
sudo reboot now
sudo apt install nginx
# proxy_pass http://127.0.0.1:1323/; と proxy_read,connect,send_timeoutを90に設定
# proxy_set_header X-Forwarded-for $remote_addr; を追加
sudo nano /etc/nginx/sites-available/passkey.tomo0611.dev
sudo ln -s /etc/nginx/sites-available/passkey.tomo0611.dev /etc/nginx/sites-enabled/passkey.tomo0611.dev
sudo systemctl restart nginx
# certbotでLet's Encryptから証明書を取得
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
