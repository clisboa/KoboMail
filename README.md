# KoboMail
Experimental email attachment downloader for Kobo devices (gmail only ATM)

## What is KoboMail?
It is a software that will read emails sent to user+kobo@gmail.com and download the attached files to the Kobo device.
It doesn't restrict the filetype so try to keep the emails you send as clean as possible ;)
The software takes advantage of a particularity in Gmail which is you can add flags to you email like the +kobo in the example above and still receive those emails on the user@gmail.com account. This makes it much easier for searching and listing emails specifically targetting your Kobo device.
Once the email is treated by program it will automatically flag it as seen so it's not processed twice (there's a flag in the config file that allows overriding for such time you want to redownload all files you've sent).

### Warning

This software is experimental, if you don't want to risk corrupting your files, database, etc of your Kobo device don't use this.
Use this software AT YOUR OWN RISK.

## Changelog
**0.1.0**
* Initial release

## Installing

Quick start for Kobo devices:

1. download the latest [KoboRoot.tgz](https://github.com/clisboa/KoboMail/releases/download/v0.1/KoboRoot.tgz)
1. connect your Kobo device to your computer with a USB cable
3. place the KoboRoot.tgz file in the .kobo directory of your Kobo device
4. disconnect the reader

When you disconnect the Kobo device, it will perform the instalation of the KoboRoot.tgz files onto the device.
Once the installation is finished you can verify that KoboRoot.tgz is now gone.
No you should head to the .add/kobomail folder and edit the kobomail_cfg.toml file


```
# currently only gmail is supported

# you need to activate IMAP for your gmail account
imap_host = "imap.gmail.com"
imap_port = "993"

# gmail account
imap_user = "user@gmail.com"

# gmail app password. you will need to generate a password specifically for KoboMail
# this can be done here: https://support.google.com/mail/answer/185833?hl=en-GB
imap_pwd = "password"

# with gmail you can send an email to user+kobo@gmail.com and the email will land on user@gmail.com account
# you can customize the flag used to detect the emails you want to process specifically for the Kobo device
email_flag = "kobo"

# flag to process all emails sent fo user+kobo@gmail.com or only the unread emails
email_unseen = "true"

# If you want to uninstall KoboMail just place an empty file called UNINSTALL next to this configuration file 
# and next time KoboMail runs it will delete itself
```

If the configuration is not correct KoboMail might not be able to work correctly.
You will need to activate IMAP on your gmail account and generate an app password as described in the config file.
Once the configuration is correct everytime your device connects to a Wifi access point the KoboMail program will run and process any emails sent to user+kobo@gmail.com that are not open yet.
If any messages were processed after a few seconds Kobo will display the dialog to connect to a PC, you don't need to actually physically connect a USB cable you just need to click on the Connect button. This is part of a workarround to trigger Kobo to recognize the new ebooks it just received via email.
After clicking on the connect button you will see the common full screen dialog as if Kobo was connected to a PC and shortly after it will show the import content progress bar.

You can attach multiple files to a single email, every attachment will be processed. All attachments will be dumped into the folder KoboMailLibrary.

There's a log.txt file in the .add/kobomail folder that will allow to diagnose problems.

## Uninstalling

Just place a file called UNINSTALL in the .add/kobomail folder and everything will be wiped clean except the KoboMailLibrary~.

## Further information.
This project includes bits and pieces of many different projects and ideas discussed in the mobileread.com forums, namely:
 - https://github.com/shermp/kobo-rclone
 - https://github.com/fsantini/KoboCloud
 - https://gitlab.com/anarcat/wallabako

