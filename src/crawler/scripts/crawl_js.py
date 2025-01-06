from bs4 import BeautifulSoup
import json, sys

req = sys.stdin.read()

def has_href_or_src(tag):
    return tag.has_attr('href') or tag.has_attr('src')
#
soup = BeautifulSoup(req, 'html.parser')
res = soup.find_all(has_href_or_src)
ret = ""
for r in res:
    try:
        ret = ret + r.attrs['src'] + "\n"
    except KeyError:
        ret = ret + r.attrs['href'] + "\n"

print(ret)
