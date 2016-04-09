import requests
import json

for i in range(0, 34):
	a = requests.get("https://api.vk.com/method/photos.get?owner_id=-23951686&album_id=wall&rev=1&extended=1&offset={}&photo_sizes=0&v=5.50".format(1000*i))
	j = json.loads(a.text)

	for v in j['response']['items']:
		print(v['photo_75'])