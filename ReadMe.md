# File Uploader

### File Uploader Server & Client

This tool provides a simple mechanism to upload files to the server using 
the standard _HTTP_ protocol.

Authentication is based on 'User - Password' pairs and restricted IP addresses.  
Use `*` as an IP address to disable IP address checks.

A pure Go language solution without any database systems or complicated set-ups.

Client interface is made with a super simple web page using a single web form.

All file uploads are verified with SHA-256 hash check sums.

## Installation
`go install github.com/vault-thirteen/File-Uploader/cmd/server@latest`

## Usage

* Create SSL certificates.
* Copy the `assets` folder.
* Create a data folder to store uploaded files.
* Modify the settings file.
  * An example of `settings.json` file can be found in the `test` folder.
* Start the server:  
`[server.exe] <Path-to-Settings-File>`

* To start the client, simply open the HTTP page with any modern web browser 
at configured address:  
`https://<Host>:<Port>`

## Why ?

Even today it can be difficult to send files between computers using different 
operating systems. Unfortunately, Linux-based systems still do not fully 
support file sharing with Microsoft Windows systems. To overcome these 
limitations, this tool was created.
