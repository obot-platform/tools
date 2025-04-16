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


if __name__ == "__main__":
    content = """
    <h2>Enhancing Operational Efficiency</h2>\n<p>AI is streamlining operations by automating routine tasks, reducing human error, and optimizing processes. Financial firms are leveraging AI capabilities to handle repetitive tasks such as data entry, document processing, and reporting. This automation not only speeds up processes but also frees employees to focus on higher-value tasks and strategic activities. 
    For instance, AI systems can process customer identity documents using optical character recognition and natural language processing, simplifying customer interactions and reducing processing time.</p>\n
    \n<img src=\\"https://testsite041stg.wpenginepowered.com/generated_image_dfb7c42a/\\" alt=\\"AI transforming finance industry\\">\n
    \n<h2>Improving Customer Experience</h2>\n
    <p>AI is playing a crucial role in delivering personalized, responsive, and convenient services at scale. Financial institutions are using AI to provide hyper-personalized experiences for each customer, analyzing factors like spending habits and life events to offer tailored suggestions and advice. AI-enabled chatbots and virtual assistants are available 24/7, handling a wide range of tasks from checking account balances to providing financial advice, thus enhancing customer satisfaction and engagement.</p>\n\n<h2>Advancing Risk Management</h2>\n<p>AI is revolutionizing risk management by analyzing large volumes of data in real-time to identify patterns and outliers that could indicate potential risks. This capability allows financial institutions to mitigate risks more accurately and swiftly than ever before. AI-powered fraud detection systems continuously monitor transactions for unusual patterns, safeguarding customer payments and reducing fraud-related losses. Additionally, AI is improving credit risk assessments by analyzing diverse data points, providing a more accurate picture of a customer's creditworthiness.</p>\n\n<h2>Driving Innovation in Financial Services</h2>\n<p>AI is fostering innovation across various facets of the finance industry. In trading, AI tools provide traders with real-time insights, optimizing trading strategies and enhancing decision-making processes. In wealth management, AI offers personalized advice and risk assessment opportunities, while in insurance, it streamlines claims processing and risk assessments. The integration of AI in these areas is creating a collaborative ecosystem that elevates the precision and effectiveness of financial services.</p>\n\n<h2>Navigating Challenges and Opportunities</h2>\n<p>Despite its transformative potential, AI integration in finance comes with challenges. Financial institutions must navigate data privacy concerns, regulatory changes, and the need for algorithmic fairness and transparency. Ensuring that AI systems operate ethically and transparently is crucial to maintaining public trust and compliance with regulations. Moreover, the industry must address cultural resistance and strategic alignment to fully leverage AI's potential.</p>\n\n<h2>The Future of AI in Finance</h2>\n<p>As AI continues to evolve, its role in finance is set to expand further. Emerging technologies like generative AI, quantum computing, and decentralized finance (DeFi) applications are poised to reshape financial services. The focus is shifting towards explainable AI, ensuring transparency and regulatory compliance in trading algorithms. Financial institutions that harness AI's potential while addressing its challenges will gain significant competitive advantages, paving the way for a more innovative, efficient, and inclusive financial landscape.</p>\n\n<p>In conclusion, AI is not just a technological upgrade but a catalyst for profound disruption in the finance industry. By embracing AI's opportunities and navigating its challenges responsibly, financial institutions can redefine their services, enhance customer experiences, and drive sustainable growth in a rapidly changing world.</p>
    """
    print(readability_metrics_tool(content))
