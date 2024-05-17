import requests
import time

# Set your GitHub API token here
token = ''

# GitHub API URL for code search
search_url = 'https://api.github.com/search/code'

# Header with your token for authentication
headers = {'Authorization': f'token {token}'}

# Parameters for the search
params = {
    'q': 'github.com/hyperledger/fabric-chaincode-go/shim in:file language:Go',
    'per_page': 100
}

def search_github(url, params, headers):
    all_items = []
    params['page'] = 1

    while True:
        response = requests.get(url, headers=headers, params=params)
        if response.status_code != 200:
            print(f"Failed to fetch data: {response.status_code}")
            break

        data = response.json()
        all_items.extend(data['items'])
        if 'next' not in response.links:
            break

        params['page'] += 1
        time.sleep(10)  # Delay to manage rate limits

    return all_items

def main():
    results = search_github(search_url, params, headers)

    with open('github_urls_2.txt', 'w') as file:
        for item in results:
            file.write(item['html_url'] + '\n')

    print(f"Total URLs fetched: {len(results)}")

if __name__ == "__main__":
    main()