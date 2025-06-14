package nginx

const nginxConfTemplate = `
# BEGIN WEBHOSTING_Hochacha {{.Username}}
	location /code/{{.Username}}/ {
    	proxy_pass http://{{.VMIP}}:8080/;
    	proxy_set_header Host $host;
    	proxy_set_header X-Real-IP $remote_addr;
	}
# END WEBHOSTING_Hochacha {{.Username}}
`

const streamConfTemplate = `
# BEGIN WEBHOSTING_STREAM_Hochacha {{.Username}}
server {
    listen {{.RandomPort}};
    proxy_pass {{.VMIP}}:22;
}
# END WEBHOSTING_STREAM_Hochacha {{.Username}}
`
