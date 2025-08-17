# KubeStellar Documentation Style Guide

_This document is just a starting point for a much more complete style guide and toolbox being developed as a CNCF LFX Mentorship project in 2025_

## Why have a Style Guide?

For KubeStellar's website and other documentation to be useful and effective, the pages need to be readable and accessible to its audience. Given the cooperative nature of an open-source project, a Style Guide makes it easier for multiple contributors to write and maintain the documentation with a consistent and understandable format and authorial voice.

An extended [Design System for KubeStellar's websites and UI is being developed](https://github.com/kubestellar/ui/blob/dev/docs/design-progress.md) during mid-2025. The resulting **Design Guide** will govern the visual language and components of KubeStellar UIs and docs.
This **Style Guide** is intended to complement the Design Guide to govern the _prose_ (written language/text) components of the KubeStellar ecosystem, and will evolve alongside the Design Guide.

## Basic Style Considerations

As a starting point, here are some very basic items to consider when working on documentation for KubeStellar and its components:

- [Use clear and concise prose](#use-clear-and-concise-prose)
- [Avoid use of emojis (especially in headings)](#avoid-use-of-emojis-in-prose)
- [Compose and format for Accessibility](#compose-and-format-for-accessibility)
- [Use care if using generative AI](#use-care-if-using-generative-ai) for writing assistance. Review text carefully to make sure it is both coherent and accurate.
- [Always check spelling and grammar](#always-do-a-spelling-and-grammar-check) before committing docs changes

### Use Clear and Concise Prose

The goal of written documentation should be "just enough." Enough detail that the reader can find the information they need, but not so much detail that they must search incessantly to find the key points.

It should also be **clear**: terms should be defined and jargon explained when first used. Long, convoluted explanations should be examined to see if they might be better broken up into smaller chunks.

Much of KubeStellar's existing documentation -- especially introductory and overview text -- is written with a slightly light and breezy tone. That is deliberate; it, however, does not mean that the texts should not be technically accurate.

### Avoid Use of Emojis in Prose

The **Design Guide** will include guidance on use and placement of iconography in KubeStellar docs and UIs.
In textual documentation, however, emojis may interfere with rendering and navigation of the website documentation, _especially_ if they are used in headings or titles. They also may make the documentation inaccessible to visitors who must use screen reader technology.

### Compose and Format for Accessibility

Writing for accessibility means ensuring that screen readers can easily read your text, your content is easy to navigate, visually clear and logically structured. This includes the appropriate usage of Heading tags, styling your text consistently and using a 'gender-neutral' language where applicable.

Accessibility is a part of any good web project, and docs are no different. As a bare minimum, we require that all our images and diagrams always have an appropriate 'alt-text' to support visually impared users and screenreaders. In Markdown, you can add alt-text using the following syntax:

```markdown
![An image of KubeStellar Contributors](images/contributors.png)
```

In this example, the text inside the square brackets [An image of KubeStellar Contributors] serves as the alt-text, and it should clearly describe the content or purpose of the image.

For a more detailed guidance on writing alt-texts, we recommend checking out [WebAIM Guide to Alternative Texts.](https://webaim.org/techniques/alttext/)

We also recommend keeping the [W3C Guidelines for Accessibility](https://www.w3.org/TR/WCAG21/) in mind when writing/illustrating KubeStellar documentation.

That said, accessibility is an ever-evolving area in technical content. As an open-source contributor, familiarizing yourself with accessibility standards will help you make more thoughtful and inclusive descisions, whether in documentation, or other areas of your work.

If you want to deepen your understanding of accessibility, we recommend checking out the following resources:

- [Gruntwork Markdown Style Guide](https://docs.gruntwork.io/guides/style/markdown-style-guide/)
- [Web Content Accessibility Guidelines (WCAG)](https://www.w3.org/WAI/standards-guidelines/wcag/)
- [Google's Style Guide for Writing Inclusive Documentation](https://developers.google.com/style/inclusive-documentation)

>**Note:** KubeStellar docs are authored primarily in **Markdown**. The examples in this guide therefore use Markdown syntax (for example, `![Alt text](path/to/image.png)`).
>While some resources above reference **HTML**, they remain useful for understanding core accessibility principles (like alt text and headings) and for cases where we embed raw HTML in Markdown. If you’re new to Markdown, start with Markdown-focused references first.

### Use Care If Using Generative AI

The use of generative AI for writing assistance is becoming more common. Any text created via generative AI tools should be **carefully** reviewed to make sure it is both coherent and accurate. Among the items to check for (this is just a start):

- Ensure the generated text is not outright plagiarism, as some LLMs have been trained on copyrighted works.
- The generated text should contain no derivative works (as understood in copyright law) of other copyrighted material.

## Always Do a Spelling and Grammar Check

Before committing and/or pushing new docs to the repository and creating a pull request, be sure to check your written content for spelling and grammar errors. This will a) prevent the need to do a secondary commit to correct any discovered during the PR review and b) avoid the risk that such an error will confuse or throw a reader out of the text when they are using the documentation.
