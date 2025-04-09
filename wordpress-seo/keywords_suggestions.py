
import re
import textstat
from collections import Counter
from sklearn.feature_extraction.text import TfidfVectorizer, ENGLISH_STOP_WORDS
import yake
import numpy as np
import openai
import os

client = openai.OpenAI(api_key=os.getenv("OPENAI_API_KEY"))


def _preprocess_text(text):
    text = re.sub(r'\s+', ' ', text)  # collapse whitespace
    text = text.strip()
    return text





# ---------- Topic Relevance via TF-IDF ----------
def extract_top_keywords_tfidf(text, top_n=10):
    text = _preprocess_text(text)
    if len(text.split()) == 0:
        return []
    vectorizer = TfidfVectorizer(stop_words='english', ngram_range=(1, 3))
    tfidf_matrix = vectorizer.fit_transform([text])
    scores = zip(vectorizer.get_feature_names_out(), tfidf_matrix.toarray()[0])
    sorted_scores = sorted(scores, key=lambda x: x[1], reverse=True)
    return sorted_scores[:top_n]

def extract_yake_keywords(text, top_n=10):
    extractor = yake.KeywordExtractor(top=top_n, stopwords=ENGLISH_STOP_WORDS)
    return extractor.extract_keywords(text)


def combined_keywords(yake_keywords, tfidf_keywords, top_n=10):
    # Normalize YAKE scores (lower is better)
    yake_scores = {kw: 1 - score for kw, score in yake_keywords}
    yake_max = max(yake_scores.values(), default=1)
    yake_norm = {kw: score / yake_max for kw, score in yake_scores.items()}

    # Normalize TF-IDF scores (higher is better)
    tfidf_scores = dict(tfidf_keywords)
    tfidf_max = max(tfidf_scores.values(), default=1)
    tfidf_norm = {kw: score / tfidf_max for kw, score in tfidf_scores.items()}

    # Combine normalized scores
    combined = {}
    for kw in set(list(yake_norm.keys()) + list(tfidf_norm.keys())):
        combined[kw] = yake_norm.get(kw, 0) + tfidf_norm.get(kw, 0)

    sorted_combined = sorted(combined.items(), key=lambda x: x[1], reverse=True)
    return sorted_combined[:top_n]


# -----------Google Suggestions----------
import requests
import pandas as pd
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics.pairwise import cosine_similarity

# Get Keyword Suggestions from Google
def get_google_suggestions(seed_keyword, lang='en', country='us'):
    url = "https://suggestqueries.google.com/complete/search"
    params = {
        "client": "firefox",
        "hl": lang,
        "gl": country,
        "q": seed_keyword
    }
    response = requests.get(url, params=params)
    response.raise_for_status()
    return response.json()[1]


def score_suggestions_by_cosine_similarity(content, suggestions):
    documents = [content] + suggestions
    vectorizer = TfidfVectorizer(stop_words='english')
    tfidf_matrix = vectorizer.fit_transform(documents)
    content_vec = tfidf_matrix[0:1]
    suggestion_vecs = tfidf_matrix[1:]
    similarities = cosine_similarity(content_vec, suggestion_vecs).flatten()
    return list(zip(suggestions, similarities))


def enrich_with_rule_scores(suggestions_with_similarity):
    enriched = []
    for keyword, score in suggestions_with_similarity:
        rule_bonus = 0
        words = keyword.split()
        if len(words) > 2:
            rule_bonus += 0.1  # bonus for long-tail keywords
        if any(w in keyword.lower() for w in ["how", "best", "guide", "tips", "ideas", "setup"]):
            rule_bonus += 0.1  # bonus for strong search intent
        final_score = score + rule_bonus
        enriched.append((keyword, score, rule_bonus, final_score))
    return enriched


def final_google_suggestions(seed_keyword, content, lang='en', country='us'):

    suggestions = get_google_suggestions(seed_keyword)
    scored = score_suggestions_by_cosine_similarity(content, suggestions)
    ranked = enrich_with_rule_scores(scored)

    df = pd.DataFrame(ranked, columns=["Keyword", "TF-IDF Similarity", "Rule Bonus", "Final Score"])
    df = df.sort_values(by="Final Score", ascending=False)
    print(df.to_markdown(index=False))
    

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
def generate_seo_keywords(content, num_keywords=10, model="gpt-4o"):
    prompt = f"""
    Given the following article content, suggest {num_keywords} SEO keywords that are highly relevant and likely to rank well. List keywords separated by commas only:

    {content} 
    """

    response = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": prompt}],
        temperature=0.1,
        max_tokens=1000
    )

    keywords_text = response.choices[0].message.content
    keywords = [kw.strip().lower() for kw in keywords_text.split(',')]
    return keywords




def keywords_suggestions(content: str):
    content = html_to_text(content)
    content = _preprocess_text(content)
    

    tfidf_keywords = extract_top_keywords_tfidf(content) 
    yake_keywords = extract_yake_keywords(content)
    combined_kw = combined_keywords(yake_keywords, tfidf_keywords)
    print("\n📌 Top Keywords (Combined):")
    for keyword, score in combined_kw:
        print(f"{keyword}: {score:.4f}")
    
    print("\n🧑‍💻 LLM Keyword Suggestions:")
    llm_keywords = generate_seo_keywords(content, num_keywords=10)
    for kw in llm_keywords:
        print(f"- {kw}")
    
    print("\n🚦 Final Keyword(s) Selection (Intersection LLM & Combined):")
    final_keywords = [kw for kw in llm_keywords if kw in dict(combined_kw)]
    print(final_keywords if final_keywords else llm_keywords[:5])  
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
    keywords_suggestions(content)