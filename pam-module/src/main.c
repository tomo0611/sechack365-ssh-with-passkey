#include <stdlib.h>
#include <stdio.h>
#include <strings.h>
#include <string.h>
#include <stdint.h>
#include <time.h>
#include <stdbool.h>
#include <security/pam_modules.h>
#include <security/pam_ext.h>
#include <sys/stat.h>
#include <unistd.h>

// Function prototypes
char *gethostnamefromfile(void);
char *generateToken(char *hostname, const char *username);
void generateQRCode(char *token, pam_handle_t *handle);
int pollingcheckAllowLogin(char *token);

// Get hostname from /etc/hostname
char *gethostnamefromfile(void)
{
	// read /etc/hostname
	FILE *fp = fopen("/etc/hostname", "r");
	if (fp == NULL)
	{
		return NULL;
	}
	// ローカル変数は関数終了時にスコープ外となり、そのメモリ領域は解放され
	// NULLになってしまうので、mallocで動的に確保する
	char *hostname = malloc(256);
	fgets(hostname, 256, fp);
	fclose(fp);
	hostname[strcspn(hostname, "\n")] = 0;
	return hostname;
}

// Generate token from authentication server
char *generateToken(char *hostname, const char *username)
{
	// ローカル変数は関数終了時にスコープ外となり、そのメモリ領域は解放され
	// NULLになってしまうので、mallocで動的に確保する
	char *token = malloc(64);
	// サーバー環境だと32だと1文字足りなかったので余裕を持って64で作成

	char str[512];
	strcpy(str, "/usr/bin/curl 'https://passkey.tomo0611.dev/api/v1/generateLoginToken");
	strcat(str, "?hostname=");
	strcat(str, hostname);
	strcat(str, "&username=");
	strcat(str, username);
	strcat(str, "'");
	// システムによっては''で囲まないとusername部分以降が送信されなかった

	FILE *fp = popen(str, "r");
	if (fp == NULL)
	{
		return NULL;
	}
	fgets(token, 32, fp);
	pclose(fp);
	return token;
}

// Generate QRCode and print it to console
void generateQRCode(char *token, pam_handle_t *handle)
{
	char str[256];
	strcpy(str, "/usr/bin/qrcode -t https://passkey.tomo0611.dev/login/");
	strcat(str, token);
	FILE *fp = popen(str, "r");
	if (fp == NULL)
	{
		printf("Failed to run command\n");
		return;
	}

	// Buffer to store each line of the file.
	char qrtext[512];
	// Read each line from the file and store it in the 'line' buffer.
	while (fgets(qrtext, sizeof(qrtext), fp))
	{
		// remove \n from the end of the line
		qrtext[strcspn(qrtext, "\n")] = 0;
		pam_info(handle, "%s", qrtext);
	}
	/* close */
	pclose(fp);
}

// Polling check allow login
int pollingcheckAllowLogin(char *token)
{
	// OK or NG
	char result[16];

	char str[256];
	strcpy(str, "/usr/bin/curl https://passkey.tomo0611.dev/api/v1/loginLongPolling/");
	strcat(str, token);

	FILE *fp = popen(str, "r");
	if (fp == NULL)
	{
		return false;
	}
	fgets(result, 16, fp);
	pclose(fp);
	if (strcmp(result, "OK") == 0)
	{
		return true;
	}
	else
	{
		return false;
	}
}

PAM_EXTERN int pam_sm_authenticate(pam_handle_t *handle, int flags, int argc,
								   const char **argv)
{
	int pam_code;

	const char *username = NULL;

	/* Asking the application for an  username */
	pam_code = pam_get_user(handle, &username, "USERNAME: ");
	if (pam_code != PAM_SUCCESS)
	{
		fprintf(stderr, "Can't get username");
		return PAM_PERM_DENIED;
	}

	// テスト用なのでuser=pskytest以外の場合は普通の認証を通す
	if (strcmp(username, "alya") != 0)
	{
		return PAM_USER_UNKNOWN;
	}

	// Get hostname
	char *hostname = gethostnamefromfile();
	if (hostname == NULL)
	{
		pam_info(handle, "Can't get hostname");
		return PAM_PERM_DENIED;
	}

	// Generate token
	char *token = generateToken(hostname, username);
	if (token == NULL)
	{
		pam_info(handle, "Can't get token, Please check the authentication server is working");
		return PAM_PERM_DENIED;
	}

	generateQRCode(token, handle);

	pam_info(handle, "\nClick https://passkey.tomo0611.dev/login/%s or scan above QR to authenticate by passkey\n", token);

	// Polling check allow login
	int allowLogin = pollingcheckAllowLogin(token);
	if (allowLogin == true)
	{
		return PAM_SUCCESS;
	}
	else
	{
		return PAM_USER_UNKNOWN;
	}
}

PAM_EXTERN int pam_sm_acct_mgmt(pam_handle_t *pamh, int flags, int argc,
								const char **argv)
{
	return PAM_SUCCESS;
}

PAM_EXTERN int pam_sm_setcred(pam_handle_t *pamh, int flags, int argc,
							  const char **argv)
{
	return PAM_SUCCESS;
}

PAM_EXTERN int pam_sm_open_session(pam_handle_t *pamh, int flags, int argc,
								   const char **argv)
{
	const char *username;
	/* Get the username from PAM */
	pam_get_item(pamh, PAM_USER, (const void **)&username);
	return PAM_SUCCESS;
	// return PAM_USER_UNKNOWN;
}

PAM_EXTERN int pam_sm_close_session(pam_handle_t *pamh, int flags, int argc,
									const char **argv)
{
	return PAM_SUCCESS;
}

PAM_EXTERN int pam_sm_chauthtok(pam_handle_t *pamh, int flags, int argc,
								const char **argv)
{
	return PAM_PERM_DENIED;
}
