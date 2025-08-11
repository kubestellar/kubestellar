# KubeStellar Documentation Style Guide

_This document is just a starting point for a much more complete style guide and toolbox being developed as a CNCF LFX Mentorship project in 2025_

## Why have a Style Guide?

For KubeStellar's website and other documentation to be useful and effective, the pages need to be readable and accessible to its audience. Given the cooperative nature of an open-source project, a Style Guide makes it easier for multiple contributors to write and maintain the documentation with a consistent and understandable format and authorial voice.

An extended [Design System for KubeStellar's websites and UI is being developed](https://github.com/kubestellar/ui/blob/dev/docs/design-progress.md) during mid-2025. The resulting **Design Guide** will govern the visual language and components of KubeStellar UIs and docs.
This **Style Guide** is intended to complement the Design Guide to govern the _prose_ (written language/text) components of the KubeStellar ecosystem, and will evolve alongside the Design Guide.

## Basic Style Considerations

As a starting point, here are some very basic items to consider when working on documentation for KubeStellar and its components:

-  [Use clear and concise prose](#use-clear-and-concise-prose)
-  [Avoid use of emojis (especially in headings)](#avoid-use-of-emojis-in-prose)
-  [Always include alt-text for images](#always-include-alt-text-for-images)
-  [Use care if using generative AI](#use-care-if-using-generative-ai) for writing assistance. Review text carefully to make sure it is both coherent and accurate.
-  [Always check spelling and grammar](#always-do-a-spelling-and-grammar-check) before committing docs changes

### Use Clear and Concise Prose

The goal of written documentation should be "just enough." Enough detail that the reader can find the information they need, but not so much detail that they must search incessantly to find the key points.

It should also be **clear**: terms should be defined and jargon explained when first used. Long, convoluted explanations should be examined to see if they might be better broken up into smaller chunks.

Much of KubeStellar's existing documentation -- especially introductory and overview text -- is written with a slightly light and breezy tone. That is deliberate; it, however, does not mean that the texts should not be technically accurate.

### Avoid Use of Emojis in Prose

The **Design Guide** will include guidance on use and placement of iconography in KubeStellar docs and UIs.
In textual documentation, however, emojis may interfere with rendering and navigation of the website documentation, _especially_ if they are used in headings or titles. They also may make the documentation inaccessible to visitors who must use screen reader technology.

### Always include alt-text for images

This is a basic requirement for accessibility. Images and diagrams should always have an appropriate alt-text attribute. In general, keep the [W3C Guidelines for Accessibility](https://www.w3.org/TR/WCAG21/) in mind when writing/illustrating KubeStellar documentation 

### Use Care If Using Generative AI

The use of generative AI for writing assistance is becoming more common. Any text created via generative AI tools should be **carefully** reviewed to make sure it is both coherent and accurate. Among the items to check for (this is just a start): 

- Ensure the generated text is not outright plagiarism, as some LLMs have been trained on copyrighted works.
- The generated text should contain no derivative works (as understood in copyright law) of other copyrighted material.

## Always Do a Spelling and Grammar Check

Before committing and/or pushing new docs to the repository and creating a pull request, be sure to check your written content for spelling and grammar errors. This will a) prevent the need to do a secondary commit to correct any discovered during the PR review and b) avoid the risk that such an error will confuse or throw a reader out of the text when they are using the documentation.