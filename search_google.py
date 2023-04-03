import requests

API_KEY = "AIzaSyDfr5L585HBytSi70XcA-8mjLLGaXMC_fE"
SEARCH_ENGINE_ID = "e0abb43f7fc734352"


# Defina suas credenciais da API do Google
api_key = API_KEY
cx = SEARCH_ENGINE_ID

# URL da imagem que você deseja pesquisar
image_url = 'https://blog.emania.com.br/wp-content/uploads/2016/02/direitos-autorais-e-de-imagem.jpg'

# URL da API de pesquisa personalizada do Google Imagens
url = f'https://www.googleapis.com/customsearch/v1?q=\
    {image_url}&cx={cx}&searchType=image&key={api_key}'

response = requests.get(url)
total_results = int(response.json()['searchInformation']['totalResults'])

# Verifica se a solicitação foi bem-sucedida
if response.status_code == 200:
    # Extrai os URLs dos resultados da pesquisa
    items = response.json().get("items", [])
    print(items)
    urls = [item.get("link") for item in items]
    print("Sites que usam esta imagem:")
    print("\n".join(urls))
else:
    print("A solicitação falhou com código de status %d" %
          response.status_code)
