from flask import Flask, request, jsonify
import os
import json
from dotenv import load_dotenv
import requests
import time

app = Flask(__name__)
app.config['JSON_SORT_KEYS'] = False

load_dotenv()

BING_KEY = os.getenv("BING_SUBSCRIPTION_KEY")
GOOGLE_KEY = os.getenv("API_KEY")
GOOGLE_ID = os.getenv("SEARCH_ENGINE_ID")
APP_KEY = os.getenv("KEY")


def bing_search(image_url, key):
    headers = {"Ocp-Apim-Subscription-Key": key}
    query_params = {
        "imgurl": image_url,
        "mkt": "en-us",
        "modules": "similarimages",
        "api-version": "7.0"
    }
    search_url = "https://api.bing.microsoft.com/v7.0/images/visualsearch"

    response = requests.post(search_url, headers=headers, params=query_params)
    response_json = response.json()

    if "error" in response_json:
        return {"error": response_json["error"]}

    results = []
    tags = response_json["tags"]
    for tag in tags:
        display_name = tag.get("displayName", "")
        if not display_name:
            actions = tag.get("actions", [])
            for action in actions:
                action_type = action.get("actionType", "")
                if action_type == "PagesIncluding":
                    values = action.get("data", {}).get("value", [])
                    for value in values:
                        name = value.get("name", "")
                        host_page_url = value.get("hostPageUrl", "")
                        content_url = value.get("contentUrl", "")
                        results.append({
                            "title": name,
                            "pageLink": host_page_url,
                            "imageLink": content_url
                        })
    return results


def get_google_url(image_url, page):
    return f'https://www.googleapis.com/customsearch/v1?q={image_url}&cx={GOOGLE_ID}&searchType=image&key={GOOGLE_KEY}&start={page}'


def google_search(image_url):

    startIndex = 1
    urls = []

    response = requests.get(get_google_url(image_url, startIndex))
    MAX_RESULTS = int(response.json()['searchInformation']['totalResults'])

    while startIndex < MAX_RESULTS:
        # Verifica se a solicitação foi bem-sucedida

        items = response.json().get("items", [])
        if (len(items) == 0):
            return urls
        print(items)
        for item in items:
            urls.append({
                "title": item["title"],
                "pageLink": item["image"]["contextLink"],
                "imageLink": item["link"]
            })
        startIndex += 10
        response = requests.get(get_google_url(image_url, startIndex))
        time.sleep(1)

    return urls


@app.route("/api/health")
def health_check():
    return "OK"


@app.route("/")
def hello_world():
    return "ONLINE"


@app.route("/api/bingsearch")
def bing_search_image():
    image_url = request.args.get("imageURL", "")
    key = request.headers.get("key", "")
    if not image_url:
        return jsonify({"error": "imageURL not provided"}), 400
    if not key or key != APP_KEY:
        return jsonify({"error": "Invalid access key" + key}), 401

    results = bing_search(image_url, BING_KEY)
    if "error" in results:
        return jsonify(results), 500

    return jsonify(results), 200


@app.route("/api/googlesearch")
def google_search_image():
    image_url = request.args.get("imageURL", "")
    key = request.headers.get("key", "")
    if not image_url:
        return jsonify({"error": "imageURL not provided"}), 400
    if not key or key != APP_KEY:
        return jsonify({"error": "Invalid access key" + key}), 401

    results = google_search(image_url)
    if "error" in results:
        return jsonify(results), 500

    return jsonify(results), 200


if __name__ == "__main__":
    port = os.getenv("PORT", 5000)
    app.run(host="0.0.0.0", port=port)
