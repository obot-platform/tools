import re
import markdownify
from tools.helper import ENGLISH_STOP_WORDS


# ---------- Tokenizer + Keyword Density ----------
def tokenize(text):
    return re.findall(r"\b\w+\b", text.lower())


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

        density = (
            round((count / total_word_count) * 100, 2) if total_word_count else 0.0
        )
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
    heading_matches = re.findall(r"^#{1,6}\s+(.*)", md_text, re.MULTILINE)
    headings = [h.strip() for h in heading_matches]

    # Remove all headings to get body
    body = re.sub(r"^#{1,6}\s+.*$", "", md_text, flags=re.MULTILINE)
    body = re.sub(r"\n{2,}", "\n", body)
    body = body.strip()

    return headings, body


def keyword_density_metrics_tool(
    title: str, content: str, primary_keyword: str, secondary_keywords: list = []
) -> dict:
    md_text = markdownify.markdownify(
        content, heading_style="ATX"
    )  # ATX: #, ##, ###, etc. This convert h2, h3, h4, etc. to ##, ###, ####, etc.

    headings, body = extract_markdown_parts(md_text)

    all_keywords = [primary_keyword] + secondary_keywords

    density_map = keyword_density(body, all_keywords)
    # evaluation = evaluate_density(density_map, primary_keyword, secondary_keywords)
    heading_presence = keyword_in_headings(headings, all_keywords)

    # --- Prepare Report Data ---
    keyword_report = []

    keyword_report.append(
        {
            "primary_keyword": primary_keyword,
            "density": round(density_map.get(primary_keyword, 0.0), 2),
            "in_title": primary_keyword.lower() in title.lower(),
            "in_heading": heading_presence.get(primary_keyword, False),
        }
    )

    for kw in secondary_keywords:
        keyword_report.append(
            {
                "secondary_keyword": kw,
                "density": round(density_map.get(kw, 0.0), 2),
                "in_heading": heading_presence.get(kw, False),
            }
        )

    return {"keyword_report": keyword_report}

