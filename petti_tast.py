#!/usr/bin/env python3

import requests
import ipaddress
import time
from concurrent.futures import ThreadPoolExecutor, as_completed

from urllib3.exceptions import InsecureRequestWarning 
requests.packages.urllib3.disable_warnings(category=InsecureRequestWarning)

def is_valid_ip(ip_str):
    try:
        ipaddress.ip_address(ip_str)
        return True
    except ValueError:
        return False

ipaddresses = [ "1.7.83.133", "101.13.1.58", "103.107.36.18", "103.115.254.109", "103.115.254.110", "103.251.31.34", "103.90.27.83", "103.99.3.226", "106.122.229.3", "109.124.215.7", "109.207.35.147", "111.194.233.62", "111.26.106.119", "111.26.95.120", "112.253.30.14", "112.30.127.9", "113.28.86.1", "115.141.143.58", "116.55.21.104", "117.159.174.136", "117.248.248.14", "121.202.206.119", "122.136.195.32", "122.185.144.150", "122.186.200.10", "122.186.243.78", "122.186.244.34", "122.187.172.98", "122.187.226.13", "123.138.101.106", "124.112.45.222", "124.198.131.35", "125.19.154.154", "125.19.206.54", "125.19.251.90", "125.20.251.66", "125.22.200.86", "125.229.102.40", "125.23.202.194", "125.64.209.11", "128.106.196.150", "128.185.187.2", "13.68.214.34", "136.232.197.106", "14.194.128.158", "149.54.15.42", "149.54.33.130", "152.52.245.38", "163.223.220.16", "171.102.130.59", "171.236.84.24", "177.0.238.190", "177.159.99.95", "178.140.212.92", "178.178.194.131", "178.57.37.17", "179.184.33.86", "182.252.140.114", "182.95.176.194", "182.95.46.186", "182.95.52.114", "182.95.52.194", "183.6.115.88", "190.143.133.126", "190.90.154.236", "192.34.164.13", "193.46.192.20", "198.91.165.166", "2.55.126.88", "209.173.10.75", "211.21.162.76", "211.216.58.204", "211.253.10.61", "213.55.85.202", "218.202.91.147", "220.182.11.126", "220.189.196.134", "220.189.253.198", "220.246.33.79", "220.246.66.209", "222.76.248.54", "27.115.42.62", "27.24.141.122", "31.41.81.65", "38.148.95.217", "39.152.138.178", "42.200.73.3", "46.149.34.226", "49.124.153.17", "49.124.159.196", "49.204.232.244", "49.249.76.221", "58.16.201.52", "58.252.212.12", "58.56.128.190", "60.223.252.57", "61.169.31.242", "61.180.116.198", "61.72.58.242", "62.182.132.94", "65.20.191.12", "65.20.251.127", "73.95.112.29", "76.177.170.161", "78.29.41.139", "81.233.235.203", "85.225.19.144", "87.12.189.79", "88.255.189.50", "89.179.78.247", "91.219.196.17", "92.126.223.175" ]

#host = "http://172.16.20.250"
host = "http://api-dev-1.ip.sunet.se"

session = requests.Session()
session.headers.update({"Accept": "application/json"})

def lookup_ip(ipaddr):
    if not is_valid_ip(ipaddr):
        return None
    try:
        resp = session.get(host + '/lookup/' + ipaddr, timeout=3)
        data = resp.json()
        if resp.status_code == 200:
            return f"{ipaddr} | CC: {data['country_iso']} | ASN: {data['asn_organization']}"
        else:
            return f"{ipaddr} | CC: (error {resp.status_code}) | ASN: (error)"
    except Exception as e:
        return f"{ipaddr} | CC: (timeout) | ASN: (timeout)"

start = time.time()

with ThreadPoolExecutor(max_workers=20) as pool:
    futures = {pool.submit(lookup_ip, ip): ip for ip in ipaddresses}
    results = {}
    for future in as_completed(futures):
        ip = futures[future]
        results[ip] = future.result()

# Print in original order
for ip in ipaddresses:
    if results.get(ip):
        print(results[ip])

print(f"\n--- {len(ipaddresses)} lookups in {time.time() - start:.2f}s ---")