import re
from collections import Counter

# from sklearn.feature_extraction.text import TfidfVectorizer
import yake
import numpy as np
import openai
import os
from tools.helper import setup_logger, ENGLISH_STOP_WORDS

logger = setup_logger(__name__)

API_KEY = os.getenv("OPENAI_API_KEY")
client = openai.OpenAI(api_key=API_KEY)
MODEL = os.getenv("OBOT_DEFAULT_LLM_MODEL", "gpt-4o")


def llm_chat_completion(messages, model=MODEL):
    return client.chat.completions.create(
        model=model, messages=messages, temperature=0.1, max_tokens=3000
    )


def _preprocess_text(text):
    text = re.sub(r"\s+", " ", text)  # collapse whitespace
    text = text.strip()
    return text


def extract_yake_keywords(text, top_n=10):
    extractor = yake.KeywordExtractor(top=top_n, stopwords=ENGLISH_STOP_WORDS)
    return extractor.extract_keywords(text)


# ---------- HTML to Text ----------
from html2text import HTML2Text


def html_to_text(html):
    converter = HTML2Text()
    converter.ignore_links = True
    converter.ignore_images = True
    converter.ignore_emphasis = False
    converter.body_width = 0  # prevent line breaks
    return converter.handle(html).strip()


# ---------- Final OpenAI Keywords Suggestion ----------
def generate_seo_keywords(content, reference_keywords, num_keywords=10, model=MODEL):
    prompt = f"""
    Given the following article content, suggest {num_keywords} SEO keywords that are highly relevant and likely to rank well. 
    Use the provided reference keywords as inspiration, but feel free to improve or modify them based on the content. 
    List the final keywords separated by commas only.

    Article Content:
    {content}

    Reference Keywords:
    {', '.join(reference_keywords)}
"""

    response = llm_chat_completion(
        messages=[{"role": "user", "content": prompt}], model=model
    )

    keywords_text = response.choices[0].message.content
    keywords = [kw.strip().lower() for kw in keywords_text.split(",")]
    return keywords


def keywords_suggestions_tool(
    content: str, num_final_keywords=5, num_generate_keywords=10
):
    content = html_to_text(content)
    content = _preprocess_text(content)

    yake_keywords = extract_yake_keywords(content)
    logger.info("\n📌 Top Keywords Yake:")
    for keyword, score in yake_keywords:
        logger.info(f"{keyword}: {score:.4f}")

    cleaned_yake_keywords = [kw for kw, _ in yake_keywords]

    logger.info("\n🧑‍💻 LLM Keyword Suggestions:")
    llm_keywords = generate_seo_keywords(
        content, cleaned_yake_keywords, num_keywords=num_generate_keywords
    )
    for kw in llm_keywords:
        logger.info(f"- {kw}")

    final_keywords = llm_keywords[:num_final_keywords]
    logger.info(final_keywords)
    return final_keywords



