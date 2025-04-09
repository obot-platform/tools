import os
import openai
from content_optimize_metrics import content_optimize_metrics


openai_api_key = os.getenv("OPENAI_API_KEY")
client = openai.OpenAI(api_key=openai_api_key)

def build_optimization_prompt( primary_keyword, secondary_keywords):
    secondary_list = ', '.join(secondary_keywords)

    prompt = f"""
You are an expert SEO content writer.

Your task is to enhance the blog post provided by the user to:

1. Enhance Readability, make it easier for 8th graders to read
- Rewrite complex or long-winded sentences where needed.
- Aim for clarity and ease of reading without changing the tone.

2. Improve Keyword Usage
- Increase the presence of the **primary keyword**: "{primary_keyword}"
- Naturally incorporate the following **secondary keywords**: {secondary_list}

Do NOT change the topic or overall structure too much — keep the content's meaning and tone.

Please return the output in HTML format with proper use of headings, paragraphs, and bullet points if applicable. Don't incldue '```html' in the output.
"""
    return prompt




def optimize_post_with_llm(content,  primary_keyword, secondary_keywords):
    prompt = build_optimization_prompt(primary_keyword, secondary_keywords)

    response = client.chat.completions.create(
        model="gpt-4o",
        messages=[{"role": "system", "content": prompt}, {"role": "user", "content": content}],
        temperature=0.7,
        max_tokens=3000
    )

    return response.choices[0].message.content.strip()


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
    import json
    primary_keyword = "gardening tips"
    secondary_keywords = ["pest control", "compost", "soil"]
    res = content_optimize_metrics(content, primary_keyword, secondary_keywords)
    
    readability_metrics = res["readability"]
    density_metrics = res["keyword_report"]
    print(f"Current Metrics:")
    print(json.dumps(res, indent=4))

    optimized_content = optimize_post_with_llm(content, primary_keyword, secondary_keywords)
    print(optimized_content)

    
    new_res = content_optimize_metrics(optimized_content, primary_keyword, secondary_keywords)
    print(f"Optimized Metrics:")
    print(json.dumps(new_res, indent=4))
    
    
