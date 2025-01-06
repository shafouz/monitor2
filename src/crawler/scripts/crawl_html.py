from bs4 import BeautifulSoup
import json, sys

req = sys.stdin.read()
filter = sys.argv[1]

soup = BeautifulSoup(req, 'html.parser')
res = soup.select(filter)

ret = []
for el in res:
    strings = list(el.strings)
    if len(strings) > 0:
        ret.extend(strings)
    else:
        if el.string != None:
            ret.append(el.string)
        elif el.text != None:
            ret.append(el.text)

print("\n".join(ret))
