#!/bin/bash

gsutil cp gs://tb_configs/appsearch/nginx/appsearch.testbook.com /etc/nginx/sites-enabled/
service nginx restart
systemctl restart enterprisesearch.service
systemctl restart hc.service

