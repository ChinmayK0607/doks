from fastapi import FastAPI, HTTPException, Request, File, UploadFile
from fastapi.responses import HTMLResponse, JSONResponse, FileResponse
from fastapi.staticfiles import StaticFiles
from pydantic import BaseModel
import sqlite3
import io
import requests
from groq import Groq
from fpdf import FPDF
from bs4 import BeautifulSoup
import re
import pytesseract
from PIL import Image

# Configuration
GROQ_API_KEY = "groq_api_key"
HF_API_URL = "https://api-inference.huggingface.co/models/Falconsai/text_summarization"
HF_API_TOKEN = "hf"  # Replace with your actual token
# pytesseract.pytesseract.tesseract_cmd = r'"C:\Tesseract\tesseract.exe"t'
app = FastAPI()

# SQLite setup
conn = sqlite3.connect('notes.db', check_same_thread=False)
c = conn.cursor()
c.execute('''CREATE TABLE IF NOT EXISTS notes
             (id INTEGER PRIMARY KEY AUTOINCREMENT,
              content TEXT)''')
conn.commit()

# Mount the static files directory
app.mount("/static", StaticFiles(directory="static"), name="static")

# Pydantic models
class NoteContent(BaseModel):
    content: str

class HTMLContent(BaseModel):
    content: str

class TextContent(BaseModel):
    text: str

# Helper functions
def convert_html_to_markdown(html_content):
    soup = BeautifulSoup(html_content, 'html.parser')
    
    for b_tag in soup.find_all('b'):
        b_tag.insert_before('**')
        b_tag.insert_after('**')
        b_tag.unwrap()
    
    for i_tag in soup.find_all('i'):
        i_tag.insert_before('*')
        i_tag.insert_after('*')
        i_tag.unwrap()
    
    markdown_content = str(soup)
    markdown_content = re.sub(r'\s+\*\*', '**', markdown_content)
    markdown_content = re.sub(r'\s+\*', '*', markdown_content)
    markdown_content = re.sub(r'\|(\*\*)', r'\1', markdown_content)
    markdown_content = re.sub(r'\|(\*)', r'\1', markdown_content)
    
    return markdown_content

def query_hf_model(payload):
    headers = {"Authorization": f"Bearer {HF_API_TOKEN}"}
    response = requests.post(HF_API_URL, headers=headers, json=payload)
    return response.json()

def expand_text_with_groq(text):
    client = Groq(api_key=GROQ_API_KEY)
    chat_completion = client.chat.completions.create(
        messages=[
            {
                "role": "user",
                "content": f"expand this further : {text}",
            }
        ],
        model="llama3-8b-8192",
    )
    return chat_completion.choices[0].message.content

# Routes
@app.get("/", response_class=HTMLResponse)
async def home():
    with open("static/index.html", "r") as file:
        content = file.read()
    return HTMLResponse(content=content)

@app.post("/notes/save")
async def save_note(note: NoteContent):
    c.execute("INSERT INTO notes (content) VALUES (?)", (note.content,))
    conn.commit()
    return {"message": "Note saved successfully"}

@app.get("/notes/list")
async def list_notes():
    c.execute("SELECT id, content FROM notes")
    notes = c.fetchall()
    return {"notes": [{"id": note[0], "content": note[1]} for note in notes]}

@app.post("/export/pdf")
async def export_pdf(request: Request):
    form = await request.form()
    content = form.get("content")

    pdf = FPDF()
    pdf.add_page()
    pdf.set_font("Arial", size=12)
    pdf.multi_cell(0, 10, content)

    pdf_buffer = io.BytesIO()
    pdf.output(pdf_buffer)
    pdf_buffer.seek(0)

    headers = {
        'Content-Disposition': 'attachment; filename=export.pdf',
        'Content-Type': 'application/pdf',
    }
    return FileResponse(pdf_buffer, headers=headers)

@app.post("/export/md")
async def export_md(request: HTMLContent):
    try:
        markdown_content = convert_html_to_markdown(request.content)
        return {"markdown": markdown_content}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/action/copy")
async def copy_to_clipboard(request: Request):
    data = await request.json()
    content = data.get("content")
    return {"message": "Copied to clipboard"}

@app.post("/action/summarize")
async def summarize_content(request: TextContent):
    try:
        output = query_hf_model({"inputs": request.text})
        summary = output[0]['summary_text']
        return {"summary": summary}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.post("/action/expand")
async def expand_content(request: TextContent):
    try:
        expansion = expand_text_with_groq(request.text)
        return {"expansion": expansion}
    except Exception as e:
        import traceback
        print("Error in /action/expand:", str(e))
        traceback.print_exc()
        raise HTTPException(status_code=500, detail=f"Expand error: {str(e)}")

@app.post("/upload/file")
async def upload_file(file: UploadFile = File(...)):
    if not file.filename.endswith('.txt'):
        raise HTTPException(status_code=400, detail="Only .txt files are allowed")
    
    try:
        contents = await file.read()
        text = contents.decode('utf-8')
        print(f"File contents: {text[:100]}...")  # Print first 100 characters for debugging
        return {"content": text, "message": "File uploaded successfully"}
    except Exception as e:
        print(f"Error reading file: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Error reading file: {str(e)}")

@app.post("/ocr")
async def ocr(file: UploadFile = File(...)):
    try:
        image_bytes = await file.read()
        image = Image.open(io.BytesIO(image_bytes))
        text = pytesseract.image_to_string(image)
        return {"text": text}
    except Exception as e:
        print(f"Error processing image: {str(e)}")
        raise HTTPException(status_code=500, detail=f"Error processing image: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    print("Server started at http://localhost:8000")
    uvicorn.run(app, host="0.0.0.0", port=5000)
