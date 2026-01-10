import { DocsLayout } from './src/components/docs/DocsLayout';
import { buildPageMap } from './src/app/docs/page-map';

// Custom MDX components without nextra-theme-docs
export function useMDXComponents(components) {
  return {
    // Wrapper component that wraps the entire MDX content with DocsLayout
    wrapper: ({ children, toc, metadata, sourceCode, ...props }) => {
      const { pageMap } = buildPageMap();
      
      return (
        <DocsLayout pageMap={pageMap} toc={toc} metadata={metadata}>
          {children}
        </DocsLayout>
      );
    },
    // You can add custom component mappings here
    // Example:
    // h1: (props) => <h1 className="custom-h1" {...props} />,
    // h2: (props) => <h2 className="custom-h2" {...props} />,
    // a: (props) => <a className="custom-link" {...props} />,
    ...components,
  };
}

