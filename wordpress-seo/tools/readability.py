import textstat
import markdownify


# ---------- Readability ----------
def ideal_readability_score():
    return {
        "flesch_reading_ease": "60-70+",
        "flesch_kincaid_grade": "7-8th grade or lower (lower is simpler)",
        "gunning_fog": "8-10 or lower (lower is simpler)",
        "smog_index": "8-10 or lower (lower is simpler)",
        "automated_readability_index": "8-10 or lower (lower is simpler)",
    }


def get_readability_scores(text: str) -> dict:
    ideal_score = ideal_readability_score()
    return {
        "flesch_reading_ease": f"score: {textstat.flesch_reading_ease(text)}, ideal_value: {ideal_score['flesch_reading_ease']}",
        "flesch_kincaid_grade": f"score: {textstat.flesch_kincaid_grade(text)}, ideal_value: {ideal_score['flesch_kincaid_grade']}",
        "gunning_fog": f"score: {textstat.gunning_fog(text)}, ideal_value: {ideal_score['gunning_fog']}",
        "smog_index": f"score: {textstat.smog_index(text)}, ideal_value: {ideal_score['smog_index']}",
        "automated_readability_index": f"score: {textstat.automated_readability_index(text)}, ideal_value: {ideal_score['automated_readability_index']}",
    }


def readability_metrics_tool(content: str) -> dict:
    md_text = markdownify.markdownify(
        content, heading_style="ATX"
    )  # ATX: #, ##, ###, etc. This convert h2, h3, h4, etc. to ##, ###, ####, etc.
    readability_metrics = get_readability_scores(md_text)
    return readability_metrics


