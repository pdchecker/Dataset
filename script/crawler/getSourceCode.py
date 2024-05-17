import base64
import requests
import os
import pandas as pd
import time
from tqdm import tqdm

# Set your GitHub API token here
token = ''

# Header with your token for authentication
headers = {'Authorization': f'token {token}'}

# get total line of code
def get_total_line(content):
    lines = content.splitlines()
    return len(lines)

def preprocess_url(url):
    parts = url.split('/')
    if len(parts) > 7 and parts[2] == 'github.com' and parts[5] == 'blob':
        parts[2] = 'api.github.com/repos'
        repo_url = '/'.join(parts[:5])
        file_path = '/'.join(parts[7:])
        code_url = f"{repo_url}/contents/{file_path}"
        return repo_url, code_url
    else:
        return None, None


def get_repo_stats(url, headers):
    """ Get information about the repository"""
    try:
        response = requests.get(url, headers=headers)
        time.sleep(1)

        if response.status_code == 200:
            repo_data = response.json()
            repo_info = {
                'RepoName': repo_data['name'],
                'Watch': repo_data['subscribers_count'],
                'Star': repo_data['stargazers_count'],
                'Fork': repo_data['forks_count']
            }
            return repo_info
        else:
            print(f"Fail to get repo {url}:", response.status_code)
            print("Err info:", response.text)
            return None
    except requests.exceptions.RequestException as e:
        print(f"Request failed on {url}:", e)
        return None


def download_code(url, headers):
    """ Download code and count lines """
    try:
        response = requests.get(url, headers=headers)
        time.sleep(1)

        if response.status_code == 200:
            content = response.json()
            code = base64.b64decode(content['content']).decode('utf-8')
            return code
        else:
            print(f"Fail to get code {url}:", response.status_code)
            print("Err info:", response.text)
            return None
    except requests.exceptions.RequestException as e:
        print(f"Request failed on {url}:", e)
        return None

def main():
    urls = [] 
    results = []
    index = 0

    with open('github_urls.txt', 'r') as file:
        urls = [line.strip() for line in file]

    for url in urls:
        repo_url, code_url= preprocess_url(url)
        code = download_code(code_url, headers)
        repo_info = get_repo_stats(repo_url, headers)
        if repo_info and code:
            file_name = f"./code/{index}.go"
            with open(file_name, 'w') as f:
                f.write(code)
            lines = get_total_line(code)
            results.append([index, repo_url, repo_info['Watch'], repo_info['Star'], repo_info['Fork'], url, lines])
            index += 1
    
    df = pd.DataFrame(results, columns=['Index', 'Repo URL', 'Watch', 'Star', 'Fork', 'Code URL', 'Line Count'])
    df.to_csv('github_code_stats.csv', index=False)

if __name__ == "__main__":
    main()