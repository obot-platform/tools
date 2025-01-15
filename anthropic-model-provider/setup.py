from setuptools import setup, find_packages

setup(
    name="anthropic_common",
    version="0.1",
    packages=find_packages(include=["anthropic_common"]),
    install_requires=["fastapi", "openai", "anthropic>=0.43.0", "openai>=1.35.7"],
)
