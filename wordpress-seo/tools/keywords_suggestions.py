
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
        model=model,
        messages=messages,
        temperature=0.1,
        max_tokens=3000
    )



def _preprocess_text(text):
    text = re.sub(r'\s+', ' ', text)  # collapse whitespace
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

    response = llm_chat_completion(messages=[{"role": "user", "content": prompt}], model=model)

    keywords_text = response.choices[0].message.content
    keywords = [kw.strip().lower() for kw in keywords_text.split(',')]
    return keywords




def keywords_suggestions_tool(content: str, num_final_keywords=5, num_generate_keywords=10):
    content = html_to_text(content)
    content = _preprocess_text(content)
    
    yake_keywords = extract_yake_keywords(content)
    logger.info("\n📌 Top Keywords Yake:")
    for keyword, score in yake_keywords:
        logger.info(f"{keyword}: {score:.4f}")
    
    cleaned_yake_keywords = [kw for kw, _ in yake_keywords]
    
    logger.info("\n🧑‍💻 LLM Keyword Suggestions:")
    llm_keywords = generate_seo_keywords(content, cleaned_yake_keywords, num_keywords=num_generate_keywords)
    for kw in llm_keywords:
        logger.info(f"- {kw}")

    final_keywords = llm_keywords[:num_final_keywords]
    logger.info(final_keywords)  
    return final_keywords

# ---------- Example Usage ----------
if __name__ == "__main__":
    content = """
<h2>Introduction</h2>
<p>Gardening is a rewarding and fulfilling activity that allows you to connect with nature, grow your own food, and create a beautiful outdoor space. Whether you&#8217;re a beginner or an experienced gardener, these essential tips will help you cultivate a thriving garden.</p>
<h2>Get to Know Your Garden</h2>
<p>Understanding your garden&#8217;s orientation and soil type is crucial. Determine if your garden is south-facing or north-facing, and test your soil to know what plants will thrive. This knowledge will guide your planting decisions.</p>
<h2>Plan Your Garden</h2>
<p>Before heading to the garden center, plan your garden layout. Consider plant compatibility and growing conditions to create a harmonious and visually appealing garden.</p>
<h2>Learn How to Plant</h2>
<p>Proper planting techniques are essential for plant health. Prepare the soil, remove weeds, and follow planting instructions to ensure your plants thrive.</p>
<h2>Feed and Water Regularly</h2>
<p>Water the root ball of plants rather than the leaves, and feed them during the growing season. Consistent care will keep your plants healthy and productive.</p>
<h2>Start Small</h2>
<p>Begin with a manageable garden size to avoid feeling overwhelmed. As you gain confidence, you can expand your garden space.</p>
<h2>Keep an Eye on Pests</h2>
<p>Monitor pest populations and take action if necessary. Natural predators can help keep pests in check, but be prepared to intervene if infestations occur.</p>
<h2>Use Compost</h2>
<p>Composting kitchen and garden waste is beneficial for the environment and your garden. Use compost as mulch to enrich the soil and promote plant growth.</p>
<h2>Prune with Confidence</h2>
<p>Pruning helps plants grow better and look good. Learn the right techniques and timing for pruning to encourage healthy growth.</p>
<h2>Be Kind to Wildlife</h2>
<p>Encourage wildlife in your garden, as they can help control pests and pollinate plants. Create habitats and learn to share your garden with them.</p>
<h2>Enjoy Your Garden</h2>
<p>Take time to relax and enjoy the space you&#8217;ve created. A garden is not just about work; it&#8217;s also a place to unwind and connect with nature.</p>
<h2>Conclusion</h2>
<p>Gardening is a journey of learning and growth. By following these tips, you&#8217;ll be well on your way to creating a beautiful and productive garden. Remember, every gardener experiences challenges, but with patience and persistence, you&#8217;ll reap the rewards of your efforts.</p>
    """
    keywords_suggestions_tool(content)