
import re
from bs4 import BeautifulSoup
import markdownify
import textstat
from collections import Counter
from sklearn.feature_extraction.text import TfidfVectorizer, ENGLISH_STOP_WORDS
import yake
import numpy as np
import os


# ---------- Readability ----------
def ideal_readability_score():
    return {
        "flesch_reading_ease": "60-70+",
        "flesch_kincaid_grade": "7-8th grade or lower (lower is simpler)",
        "gunning_fog": "8-10 or lower (lower is simpler)",
        "smog_index": "8-10 or lower (lower is simpler)",
        "automated_readability_index": "8-10 or lower (lower is simpler)"
    }


def get_readability_scores(text):
    ideal_score = ideal_readability_score()
    return {
        "flesch_reading_ease": f"score: {textstat.flesch_reading_ease(text)}, ideal_value: {ideal_score['flesch_reading_ease']}",
        "flesch_kincaid_grade": f"score: {textstat.flesch_kincaid_grade(text)}, ideal_value: {ideal_score['flesch_kincaid_grade']}",
        "gunning_fog": f"score: {textstat.gunning_fog(text)}, ideal_value: {ideal_score['gunning_fog']}",
        "smog_index": f"score: {textstat.smog_index(text)}, ideal_value: {ideal_score['smog_index']}",
        "automated_readability_index": f"score: {textstat.automated_readability_index(text)}, ideal_value: {ideal_score['automated_readability_index']}"
    }



# ---------- Tokenizer + Keyword Density ----------
def tokenize(text):
    return re.findall(r'\b\w+\b', text.lower())

def keyword_density(text, target_keywords):
    """
    Compute keyword density for selected keywords.

    Args:
        text (str): The content to analyze.
        target_keywords (list of str): Keywords or phrases to measure.

    Returns:
        dict: {keyword: density_percentage}
    """
    words = tokenize(text)
    total_word_count = len([w for w in words if w not in ENGLISH_STOP_WORDS])

    # Normalize content into lowercase phrases for multi-word matching
    text_lower = text.lower()

    density_scores = {}
    for keyword in target_keywords:
        # Count keyword frequency (exact phrase match, case-insensitive)
        # For multi-word keywords, count overlapping occurrences
        pattern = re.escape(keyword.lower())
        matches = re.findall(pattern, text_lower)
        count = len(matches)

        density = round((count / total_word_count) * 100, 2) if total_word_count else 0.0
        density_scores[keyword] = density

    return density_scores

# ---------- Density Evaluation (n-gram aware) ----------
def evaluate_density(density_map, primary_keyword, secondary_keywords):
    def bounds(keyword):
        n = len(keyword.strip().split())
        if n == 1:
            return 1.0, 2.5
        elif n == 2:
            return 0.5, 1.5
        else:
            return 0.1, 1.0

    eval_map = {}
    for kw, pct in density_map.items():
        min_thres, max_thres = bounds(kw)
        if kw == primary_keyword or kw in secondary_keywords:
            eval_map[kw] = "OK" if min_thres <= pct <= max_thres else "Adjust"
        else:
            eval_map[kw] = "Not prioritized"
    return eval_map


# --- Keyword in Headings ---
def keyword_in_headings(headings, keywords):
    result = {}
    for kw in keywords:
        in_headings = any(kw.lower() in h.lower() for h in headings)
        result[kw] = in_headings
    return result



# --- Extract from Markdown ---
def extract_markdown_parts(md_text: str):
    # Headings: lines starting with #, ##, ###, etc.
    heading_matches = re.findall(r'^#{1,6}\s+(.*)', md_text, re.MULTILINE)
    headings = [h.strip() for h in heading_matches]

    # Remove all headings to get body
    body = re.sub(r'^#{1,6}\s+.*$', '', md_text, flags=re.MULTILINE)
    body = re.sub(r'\n{2,}', '\n', body) 
    body = body.strip()

    return headings, body


def content_optimize_metrics(content: str, primary_keyword: str, secondary_keywords: list = []) -> dict:
    md_text = markdownify.markdownify(content, heading_style="ATX") # ATX: #, ##, ###, etc. This convert h2, h3, h4, etc. to ##, ###, ####, etc.
    readability_metrics = get_readability_scores(md_text)

    headings, body = extract_markdown_parts(md_text)

    all_keywords = [primary_keyword] + secondary_keywords

    density_map = keyword_density(body, all_keywords)
    evaluation = evaluate_density(density_map, primary_keyword, secondary_keywords)
    heading_presence = keyword_in_headings(headings, all_keywords)

    # --- Prepare Report Data ---
    keyword_report = []
    for kw in all_keywords:
        keyword_report.append({
            "keyword": kw,
            "density": round(density_map.get(kw, 0.0), 2),
            "in_heading": heading_presence.get(kw, False),
            "evaluation": evaluation.get(kw, "–")
        })

    return {
        "readability": readability_metrics,
        "keyword_report": keyword_report
    }



# ---------- Main ----------
if __name__ == "__main__":
    # Sample blog content
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

    primary_keyword = "gardening tips"
    secondary_keywords = ["pest control", "compost", "soil"]
    res = content_optimize_metrics(content, primary_keyword, secondary_keywords)
    import json
    print(json.dumps(res, indent=4))