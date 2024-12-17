/**
 * @fileoverview Shared types for ESLint Core.
 */
/**
 * Represents an error inside of a file.
 */
export interface FileError {
    message: string;
    line: number;
    column: number;
    endLine?: number;
    endColumn?: number;
}
/**
 * Represents a problem found in a file.
 */
export interface FileProblem {
    ruleId: string | null;
    message: string;
    loc: SourceLocation;
}
/**
 * Represents an AST node or token with location information in ESLint format.
 */
export interface SyntaxElement {
    loc: SourceLocation;
    range: SourceRange;
}
/**
 * Represents the start and end coordinates of a node inside the source.
 */
export interface SourceLocation {
    start: Position;
    end: Position;
}
/**
 * Represents the start and end coordinates of a node inside the source with an offset.
 */
export interface SourceLocationWithOffset {
    start: PositionWithOffset;
    end: PositionWithOffset;
}
/**
 * Represents a location coordinate inside the source. ESLint-style formats
 * have just `line` and `column` while others may have `offset` as well.
 */
export interface Position {
    line: number;
    column: number;
}
/**
 * Represents a location coordinate inside the source with an offset.
 */
export interface PositionWithOffset extends Position {
    offset: number;
}
/**
 * Represents a range of characters in the source.
 */
export type SourceRange = [number, number];
/**
 * The human readable severity level used in a configuration.
 */
export type SeverityName = "off" | "warn" | "error";
/**
 * The numeric severity level for a rule.
 *
 * - `0` means off.
 * - `1` means warn.
 * - `2` means error.
 *
 */
export type SeverityLevel = 0 | 1 | 2;
/**
 * The severity of a rule in a configuration.
 */
export type Severity = SeverityName | SeverityLevel;
/**
 * Represents the configuration options for the core linter.
 */
export interface LinterOptionsConfig {
    /**
     * Indicates whether or not inline configuration is evaluated.
     */
    noInlineConfig?: boolean;
    /**
     * Indicates what to do when an unused disable directive is found.
     */
    reportUnusedDisableDirectives?: boolean | Severity;
}
/**
 * Shared settings that are accessible from within plugins.
 */
export type SettingsConfig = Record<string, unknown>;
/**
 * The configuration for a rule.
 */
export type RuleConfig = Severity | [Severity, ...unknown[]];
/**
 * A collection of rules and their configurations.
 */
export type RulesConfig = Record<string, RuleConfig>;
/**
 * Represents a plugin language.
 */
export interface Language {
    /**
     * Indicates how ESLint should read the file.
     */
    fileType: "text";
    /**
     * First line number returned from the parser (text mode only).
     */
    lineStart: 0 | 1;
    /**
     * First column number returned from the parser (text mode only).
     */
    columnStart: 0 | 1;
    /**
     * The property to read the node type from. Used in selector querying.
     */
    nodeTypeKey: string;
    /**
     * The traversal path that tools should take when evaluating the AST
     */
    visitorKeys?: Record<string, string[]>;
    /**
     * Validates languageOptions for this language.
     */
    validateLanguageOptions(languageOptions: LanguageOptions): void;
    /**
     * Helper for esquery that allows languages to match nodes against
     * class. esquery currently has classes like `function` that will
     * match all the various function nodes. This method allows languages
     * to implement similar shorthands.
     */
    matchesSelectorClass?(className: string, node: object, ancestry: object[]): boolean;
    /**
     * Parses the given file input into its component parts. This file should not
     * throws errors for parsing errors but rather should return any parsing
     * errors as parse of the ParseResult object.
     */
    parse(file: File, context: LanguageContext): ParseResult;
    /**
     * Creates SourceCode object that ESLint uses to work with a file.
     */
    createSourceCode(file: File, input: OkParseResult, context: LanguageContext): SourceCode;
}
/**
 * Plugin-defined options for the language.
 */
export type LanguageOptions = Record<string, unknown>;
/**
 * The context object that is passed to the language plugin methods.
 */
export interface LanguageContext {
    languageOptions: LanguageOptions;
}
/**
 * Represents a file read by ESLint.
 */
export interface File {
    /**
     * The path that ESLint uses for this file. May be a virtual path
     * if it was returned by a processor.
     */
    path: string;
    /**
     * The path to the file on disk. This always maps directly to a file
     * regardless of whether it was returned from a processor.
     */
    physicalPath: string;
    /**
     * Indicates if the original source contained a byte-order marker.
     * ESLint strips the BOM from the `body`, but this info is needed
     * to correctly apply autofixing.
     */
    bom: boolean;
    /**
     * The body of the file to parse.
     */
    body: string | Uint8Array;
}
/**
 * Represents the successful result of parsing a file.
 */
export interface OkParseResult<T extends object = object> {
    /**
     * Indicates if the parse was successful. If true, the parse was successful
     * and ESLint should continue on to create a SourceCode object and run rules;
     * if false, ESLint should just report the error(s) without doing anything
     * else.
     */
    ok: true;
    /**
     * The abstract syntax tree created by the parser. (only when ok: true)
     */
    ast: T;
    /**
     * Any additional data that the parser wants to provide.
     */
    [key: string]: any;
}
/**
 * Represents the unsuccessful result of parsing a file.
 */
export interface NotOkParseResult {
    /**
     * Indicates if the parse was successful. If true, the parse was successful
     * and ESLint should continue on to create a SourceCode object and run rules;
     * if false, ESLint should just report the error(s) without doing anything
     * else.
     */
    ok: false;
    /**
     * Any parsing errors, whether fatal or not. (only when ok: false)
     */
    errors: FileError[];
    /**
     * Any additional data that the parser wants to provide.
     */
    [key: string]: any;
}
export type ParseResult<T extends object = object> = OkParseResult<T> | NotOkParseResult;
/**
 * Represents inline configuration found in the source code.
 */
interface InlineConfigElement {
    /**
     * The location of the inline config element.
     */
    loc: SourceLocation;
    /**
     * The interpreted configuration from the inline config element.
     */
    config: {
        rules: RulesConfig;
    };
}
/**
 * Represents the basic interface for a source code object.
 */
interface SourceCodeBase {
    /**
     * Root of the AST.
     */
    ast: object;
    /**
     * The traversal path that tools should take when evaluating the AST.
     * When present, this overrides the `visitorKeys` on the language for
     * just this source code object.
     */
    visitorKeys?: Record<string, string[]>;
    /**
     * Retrieves the equivalent of `loc` for a given node or token.
     */
    getLoc(nodeOrToken: object): SourceLocation;
    /**
     * Retrieves the equivalent of `range` for a given node or token.
     */
    getRange(nodeOrToken: object): SourceRange;
    /**
     * Traversal of AST.
     */
    traverse(): Iterable<TraversalStep>;
    /**
     * Applies language options passed in from the ESLint core.
     */
    applyLanguageOptions?(languageOptions: LanguageOptions): void;
    /**
     * Return all of the inline areas where ESLint should be disabled/enabled
     * along with any problems found in evaluating the directives.
     */
    getDisableDirectives?(): {
        directives: Directive[];
        problems: FileProblem[];
    };
    /**
     * Returns an array of all inline configuration nodes found in the
     * source code.
     */
    getInlineConfigNodes?(): object[];
    /**
     * Applies configuration found inside of the source code. This method is only
     * called when ESLint is running with inline configuration allowed.
     */
    applyInlineConfig?(): {
        configs: InlineConfigElement[];
        problems: FileProblem[];
    };
    /**
     * Called by ESLint core to indicate that it has finished providing
     * information. We now add in all the missing variables and ensure that
     * state-changing methods cannot be called by rules.
     * @returns {void}
     */
    finalize?(): void;
}
/**
 * Represents the source of a text file being linted.
 */
export interface TextSourceCode extends SourceCodeBase {
    /**
     * The body of the file that you'd like rule developers to access.
     */
    text: string;
}
/**
 * Represents the source of a binary file being linted.
 */
export interface BinarySourceCode extends SourceCodeBase {
    /**
     * The body of the file that you'd like rule developers to access.
     */
    body: Uint8Array;
}
export type SourceCode = TextSourceCode | BinarySourceCode;
/**
 * Represents a traversal step visiting the AST.
 */
export interface VisitTraversalStep {
    kind: 1;
    target: unknown;
    phase: 1 | 2;
    args: unknown[];
}
/**
 * Represents a traversal step calling a function.
 */
export interface CallTraversalStep {
    kind: 2;
    target: string;
    phase?: string;
    args: unknown[];
}
export type TraversalStep = VisitTraversalStep | CallTraversalStep;
/**
 * The type of disable directive. This determines how ESLint will disable rules.
 */
export type DirectiveType = "disable" | "enable" | "disable-line" | "disable-next-line";
/**
 * Represents a disable directive.
 */
export interface Directive {
    /**
     * The type of directive.
     */
    type: DirectiveType;
    /**
     * The node of the directive. May be in the AST or a comment/token.
     */
    node: unknown;
    /**
     * The value of the directive.
     */
    value: string;
    /**
     * The justification for the directive.
     */
    justification?: string;
}
export {};
