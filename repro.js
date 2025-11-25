const content = `
Some text.
{/Included in website. Edit CONTRIBUTING.md for GitHub./}
<!-- Included in website. Edit CONTRIBUTING.md for GitHub. -->
{/Canonical GitHub version. Edit contributing-inc.md for website./}
<!-- Canonical GitHub version. Edit contributing-inc.md for website. -->
{/A wrapper file to include the GOVERNANCE file from the repository root/}
<!-- A wrapper file to include the GOVERNANCE file from the repository root -->
{/Code management Prow, Gh actions broken links, pr verifier, emoji in titles of prs, add issue to project. Add pr to project. Check spelling errors, wordlist.txt, Quay.io/}
<!-- Code management Prow, Gh actions broken links, pr verifier, emoji in titles of prs, add issue to project. Add pr to project. Check spelling errors, wordlist.txt, Quay.io -->
End text.
`;

function removeCommentPatterns(content) {
  let cleaned = content;
  cleaned = cleaned.replace(
    /\{\/Included in website\. Edit CONTRIBUTING\.md for GitHub\.\/\}/gi,
    "REMOVED_MDX_1"
  );
  cleaned = cleaned.replace(
    /<!--\s*Included in website\. Edit CONTRIBUTING\.md for GitHub\.\s*-->/gi,
    "REMOVED_HTML_1"
  );

  cleaned = cleaned.replace(
    /\{\/Canonical GitHub version\. Edit contributing-inc\.md for website\.\/\}/gi,
    "REMOVED_MDX_2"
  );
  cleaned = cleaned.replace(
    /<!--\s*Canonical GitHub version\. Edit contributing-inc\.md for website\.\s*-->/gi,
    "REMOVED_HTML_2"
  );

  cleaned = cleaned.replace(
    /\{\/A wrapper file to include the GOVERNANCE file from the repository root\/\}/gi,
    "REMOVED_MDX_3"
  );
  cleaned = cleaned.replace(
    /<!--\s*A wrapper file to include the GOVERNANCE file from the repository root\s*-->/gi,
    "REMOVED_HTML_3"
  );

  cleaned = cleaned.replace(
    /\{\/Code management Prow[\s\S]*?Quay\.io\/\}/gi,
    "REMOVED_MDX_4"
  );
  cleaned = cleaned.replace(
    /<!--\s*Code management Prow[\s\S]*?Quay\.io\s*-->/gi,
    "REMOVED_HTML_4"
  );

  return cleaned;
}

const cleaned = removeCommentPatterns(content);
console.log(cleaned);
