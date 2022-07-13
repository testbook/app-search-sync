server {
        listen 80;
        server_name appsearch.testbook.com www.appsearch.testbook.com;
        location / {

        proxy_set_header        Host $host;
        proxy_set_header        X-Real-IP $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header        X-Forwarded-Proto $scheme;

        proxy_pass          	http://localhost:3002;
	error_log		/var/log/nginx/search_error.log;

        }

#        location  /as/ {
#
#	auth_basic "authentication required";
#	      	auth_basic_user_file /etc/nginx/.htpasswd;
#
#        proxy_set_header        Host $host;
#        proxy_set_header        X-Real-IP $remote_addr;
#        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
#        proxy_set_header        X-Forwarded-Proto $scheme;
#
#        proxy_pass          	http://localhost:3002;
#	error_log		/var/log/nginx/search_error_as.log;
#
#        }


        location ^~ /health {
        proxy_set_header        Host $host;
        proxy_set_header        X-Real-IP $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header        X-Forwarded-Proto $scheme;
        proxy_pass          	http://localhost:3000;
	 error_log		/var/log/nginx/health_error.log;
	}
}
server {
	listen 443 ssl;
        server_name appsearch.testbook.com www.appsearch.testbook.com;
        ssl_protocols TLSv1 TLSv1.1 TLSv1.2;
        # done from https://weakdh.org/sysadmin.html
        ssl_ciphers 'ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA';
        ssl_prefer_server_ciphers on;

	ssl_certificate /home/testbook/new_certs/new_2022/final.crt;
	ssl_certificate_key /home/testbook/new_certs/new_2022/testbook_ssl_rsa;
	ssl_dhparam /home/testbook/new_certs/new_2022/dhparams.pem;

	error_log /var/log/nginx/appsearch.log;

        location / {

        proxy_set_header        Host $host;
        proxy_set_header        X-Real-IP $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header        X-Forwarded-Proto $scheme;

        proxy_pass          	http://localhost:3002;
	error_log		/var/log/nginx/search_error.log;

        }

#        location  /as/ {
#
#	auth_basic "authentication required";
#	      	auth_basic_user_file /etc/nginx/.htpasswd;
#
#        proxy_set_header        Host $host;
#        proxy_set_header        X-Real-IP $remote_addr;
#        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
#        proxy_set_header        X-Forwarded-Proto $scheme;
#
#        proxy_pass          	http://localhost:3002;
#	error_log		/var/log/nginx/search_error_as.log;
#
#        }


        location ^~ /health {
        proxy_set_header        Host $host;
        proxy_set_header        X-Real-IP $remote_addr;
        proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header        X-Forwarded-Proto $scheme;
        proxy_pass          	http://localhost:3000;
	 error_log		/var/log/nginx/health_error.log;
	}
}
