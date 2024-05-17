import os
import pandas as pd

def check_keyword_in_file(file_path, keyword):
    """Check if the file contains specific keywords"""
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            return keyword in file.read()
    except:
        return False

def analyze_files_in_directory(directory_path, keyword):
    """record keyword occurrences"""
    data = {'File Name': [], 'Contains Keyword': []}
    for root, dirs, files in os.walk(directory_path):
        for file in files:
            file_path = os.path.join(root, file)
            file_name_without_extension = os.path.splitext(file)[0]
            contains_keyword = check_keyword_in_file(file_path, keyword)
            data['File Name'].append(file_name_without_extension)
            data['Contains Keyword'].append(1 if contains_keyword else 0)
    return pd.DataFrame(data)

directory_path = './code' 
keyword = 'PrivateData'
result_df = analyze_files_in_directory(directory_path, keyword)

result_df.to_csv('PDC.csv', index=False)