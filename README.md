# Dataset of Chaincode & StackOverflow Posts

- This is the dataset for the paper "Understanding and Detecting Privacy Leakage Vulnerabilities in Hyperledger Fabric Chaincodes"
- It includes 2,000 chaincodes from GitHub, 263 posts from StackOverflow, as well as the analysis and evaluation results for these data.

The directory structure is shown below:

  Dataset
  ├── README.md
  ├── chaincode
  ├── results
  │   ├── codeInfo.xlsx
  │   ├── detectionOutput
  │   └── stackoverflowQA.xlsx
  └── script
      ├── crawler
      │   ├── codeLinks.txt
      │   ├── getCodeLink.py
      │   └── getSourceCode.py
      └── process
          └── isApplyingPDC.py

---

## Chaincode

- All chaincodes are developed in Go language.
  - 0-999 are retrieved by keyword "github.com/hyperledger/fabric-contract-api-go"
  - 1,000-1,999 are retrieved by keyword "github.com/hyperledger/fabric-chaincode-go"

## Result

- **stackoverflowQA.xlsx**: information of 263 labeled StackOverflow posts.
- **detectionOutput**: detection result of PDChecker on collected chaincodes against privacy leakage vulnerabilities.
- **codeInfo.xlsx**: detailed information and distribution of privacy leakage vulnerabilities in each chaincode.

## Script

### crawler

- **getCodeLink.py**: crawling chaincode links via GitHub API.
- **getSourceCode.py**: crawling souce code via chaincode links.

### process

- **isApplyingPDC.py**: checking if PDC is applied in chaincodes.
