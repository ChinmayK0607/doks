from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel
from bs4 import BeautifulSoup
import re
import requests

app = FastAPI()

class HTMLContent(BaseModel):
    content: str

class TextContent(BaseModel):
    text: str

def convert_html_to_markdown(html_content):
    # Parse the HTML content
    soup = BeautifulSoup(html_content, 'html.parser')

    # Convert <b> tags to **
    for b_tag in soup.find_all('b'):
        b_tag.insert_before('**')
        b_tag.insert_after('**')
        b_tag.unwrap()

    # Convert <i> tags to *
    for i_tag in soup.find_all('i'):
        i_tag.insert_before('*')
        i_tag.insert_after('*')
        i_tag.unwrap()

    # Convert the soup object back to string
    markdown_content = str(soup)

    # Remove spaces before closing tags and convert closing tags
    markdown_content = re.sub(r'\s+\*\*', '**', markdown_content)  # Remove spaces before closing **
    markdown_content = re.sub(r'\s+\*', '*', markdown_content)    # Remove spaces before closing *
    # Fix ** and * issues
    markdown_content = re.sub(r'\|(\*\*)', r'\1', markdown_content)
    markdown_content = re.sub(r'\|(\*)', r'\1', markdown_content)

    return markdown_content

def query_hf_model(payload):
    API_URL = "https://api-inference.huggingface.co/models/Falconsai/text_summarization"
    headers = {"Authorization": "Bearer hf_----"} # add your hf_token
    response = requests.post(API_URL, headers=headers, json=payload)
    return response.json()

@app.post("/export/md")
async def export_md(request: HTMLContent):
    try:
        markdown_content = convert_html_to_markdown(request.content)
        return {"markdown": markdown_content}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/summarize")
async def summarize(request: TextContent):
    try:
        output = query_hf_model({"inputs": request.text})
        summary = output[0]['summary_text']
        return {"summary": summary}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/")
async def read_root():
    return {"message": "Hello World"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
