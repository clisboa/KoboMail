# KoboMail
Experimental email attachment downloader for Kobo devices

## What is KoboMail?
It is a software that will read emails sent to an email account and download the attached files to the Kobo device.
Once the email is treated by program it will automatically flag it as seen so it's not processed twice (there's a flag in the config file that allows overriding for such time you want to redownload all files you've sent).

### Warning

This software is experimental, if you don't want to risk corrupting your files, database, etc of your Kobo device don't use this.
Use this software AT YOUR OWN RISK.

## Changelog
**0.3.0**
* added hability to configure email search criteria by subject tag
* added hability to limit which filetypes we get from the emails found
* users can choose to run KoboMail automatically via UDEV rules (whenever Kobo device gets wifi) or manually via NickelMenu
* NickelDbus usage now requires new version 0.2.0 for dialog (if not present only logs will be written)

**0.2.0**
* Added optional dependency to NickelDbus so we can trigger library rescan without human intervention
* with NickelDbus the user now gets a toast when no new emails were processed and a dialog showing the number of ebooks download and processed (only if NickelDbus is present, else the old method still aplies)

![User Dialog](https://clisboa.github.io/img/KoboMailDialog.jpg)

**0.1.0**
* Initial release


## TODO
* refactoring NickelDbus to use DBUs libray instead of the qndb binary
* add kepubify (already verified the arm binary can perfectly do the conversion on device)
* find latest versions of NickelDbus and NickelMenu and propose to update that for the user
* ~~looking into using NickelDbus to replace udev rules (triggering the program execution)~~
* REMOVED - NickelDBus is a better alternative ~~* adding fbink progress bar when we start execution? (probably to much hassle for little return in functionality)~~
* ~~add other emails accounts types~~

## Installing

Quick start for Kobo devices:
1. (OPTIONAL) download and install NickelDbus for increased functionality [KoboRoot.tgz](https://github.com/shermp/NickelDBus/releases/download/0.2.0/KoboRoot.tgz)
2. (OPTIONAL) download and install NickelMenu for increased functionality [KoboRoot.tgz](https://github.com/pgaskin/NickelMenu/releases/download/v0.5.2/KoboRoot.tgz)
3. download the latest [KoboRoot.tgz](https://github.com/clisboa/KoboMail/releases/download/v0.3/KoboRoot.tgz)
4. connect your Kobo device to your computer with a USB cable
5. place the KoboRoot.tgz file in the .kobo directory of your Kobo device
6. disconnect the reader

When you disconnect the Kobo device, it will perform the instalation of the KoboRoot.tgz files onto the device.
Once the installation is finished you can verify that KoboRoot.tgz is now gone.
No you should head to the .adds/kobomail folder and edit the kobomail_cfg.toml file


```
# If you want to uninstall KoboMail just place an empty file called UNINSTALL next to this configuration file 
# and next time KoboMail runs it will delete itself

[imap_config]
    # you need to activate IMAP for your gmail account
    imap_host = "imap.gmail.com"
    imap_port = "993"

    # email account
    imap_user = "user@gmail.com"

    # for gmail you should use the following instructions
    # gmail app password. you will need to generate a password specifically for KoboMail
    # this can be done here: https://support.google.com/mail/answer/185833?hl=en-GB
    # other email services please review their configuration options
    imap_pwd = "password"

    # there's two methods KoboMail can identify emails destined to be imported into you're Kobo device:
    #  - plus:      where the email server allows sending emails to user+flag@server.com (like gmail)
    #email_flag_type = "plus"
    #email_flag = "kobo"

    #  - subject:   where KoboMail will search for emails with a subject like [flag] or $flag$ . 
    #               We recommend using something unique so there's no false positives, for example
    #               if you just put Kobo you might be allowing KoboMail to detect regular emails from Rakuten Kobo
    email_flag_type = "subject"
    email_flag = "[MyKobo]"

    # flag to process all emails sent to kobo device or only the unread emails 
    email_unseen = "true"

[execution_type]
    # there's two methods to trigger the KoboMail execution:
    #  - auto:   via a trigger attached to the wifi connection, where whenever your kobo device is connected to a wifi connection KoboMail is run
    type = "auto"
    #  - manual: via NickelMenu where your Kobo device receives an adicional menu where you can mannually trigger the KoboMail execution
    #             for this method to work you should first install NickelMenu, head over to https://pgaskin.net/NickelMenu/ to review this component 
    #type = "manual"

[processing_config]
    #list the files KoboMail should get from the emails:
    #filetypes = ["epub", "kepub", "mobi", "pdf", "cbz", "cbr", "txt", "rtf"]
    filetypes = ["epub", "kepub"]

    #process epub files with kepubify to generate the kepub version with know improvements
    #not yet implemented
    #kepubify = "true"
```

If the configuration is not correct KoboMail might not be able to work correctly.
Currently KoboMail will allow accessing any imap email server, altough tests have been done only in gmail.
The search criteria can be defined in the configuration file and there's two methods:
- plus: where KoboMail will find emails sent to user+tag@server.com
- subject: where KoboMail will search emails sent to user@server.com with the [MyKobo] tab in the subject

There's also two different methods to start KoboMail:
- manual: if you have NickelMenu installed there will be a new option to start KoboMail manually
- auto: via a automatic mechanism (UDEV rules) which will trigger KoboMail everytime Wifi is activated

The prefered method is "manual" as probably you don't want to start KoboMail everytime you sync data.

Everytime KoboMail connects and finds new ebooks to be added the import screen will be shown and the new ebooks will be available in My Books section.
You might want to review the filetypes allowed by default, currently only kepub and epub.

You can attach multiple files to a single email, every attachment will be processed. All attachments will be dumped into the folder KoboMailLibrary.

There's a kobomail.log file in the .adds/kobomail folder that will allow to diagnose problems.

## Uninstalling

Just place a file called UNINSTALL in the .adds/kobomail folder and everything will be wiped clean except the KoboMailLibrary.

## Further information.
This project includes bits and pieces of many different projects and ideas discussed in the mobileread.com forums, namely:
 - https://github.com/shermp/kobo-rclone
 - https://github.com/fsantini/KoboCloud
 - https://gitlab.com/anarcat/wallabako
 - https://github.com/shermp/NickelDBus

