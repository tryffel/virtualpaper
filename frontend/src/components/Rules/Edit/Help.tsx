import { HelpButton } from "../../Help";
import { Typography } from "@mui/material";
import * as React from "react";

export const RuleEditHelp = () => {
  const dateRegexExample = `(\d{4}-\d{1,2}-\d{1,2})`;

  return (
    <HelpButton title="Edit Rule">
      <Typography variant="h6" color="textPrimary">
        What are processing rules
      </Typography>
      <p>
        Processing rules are a set of instructions that try to minimize the
        manual work when uploading documents by automatically detecting document
        and modifying its contents. The idea is to configure a set of conditions
        that must match, and running a set of actions for documents that match
        the conditions.
      </p>
      <p>
        Processing rules can try to detect document content, name, description
        or metadata and modify the same fields.
      </p>
      <p>
        If e.g. Google regularly sends invoices, it might make sense to try to
        detect documents where:
        <ol>
          <li>Sender is Google</li>
          AND
          <li>Document is an invoice</li>
        </ol>
        If so, then:
        <ol>
          <li>Set name to 'Google monthly invoice'</li>
          <li>Add metadata 'class:invoice'</li>
          <li>Add metadata 'company:google'</li>
        </ol>
      </p>
      <Typography variant="h5" color="textPrimary">
        Instructions
      </Typography>
      <p>
        Match conditions:
        <ul>
          <li>Match all: all conditions must match</li>
          <li>Match any: any condition must match</li>
        </ul>
      </p>
      <p>
        Condition settings
        <ul>
          <li>
            Enabled: user can toggle each condition on and off without deleting
            it
          </li>
          <li>
            Case insensitive: whether to match text in case insensitive. Only
            applies to name, description or content filters
          </li>
          <li>
            Inverted: boolean negation. If selected, the condition result is
            negated.
          </li>
          <li>
            Regex: whether or not the filter is a regular expression. Only
            applies to name, description or content filters.
          </li>
        </ul>
      </p>
      <p>
        Action settings
        <ul>
          <li>
            Enabled: user can toggle each action on and off without deleting it.
          </li>
          <li>
            On condition: if true, action is only executed when conditions are
            met. If false, action is executed if conditions are not met
          </li>
        </ul>
      </p>
      <Typography variant="h6">Extracting date</Typography>
      <p>
        Matching a date from the document is a special case of condition. By
        setting condition type to 'date is' the automation searches for dates
        inside the document. The automation searches for the given regular
        expressions to match date. If date is found, then the date time is
        extracted using the date format. Thus regular expression controls
        finding the date time text inside the document and date format controls
        how the matched date string is converted to date time. For more info on
        possible time formats, see Golang's documentation on time formats:
        <a href="https://pkg.go.dev/time#pkg-constants">pkg.go.dev/time</a>
      </p>
      <p>
        Fox configuring date extraction set following settings:
        <ol>
          <li>Set regex to true</li>
          <li>Enter regular expression</li>
          <li>
            Enter a valid date time format as per Golang time parsing formats.{" "}
          </li>
        </ol>
        Example values could be:
        <ol>
          <li>filter: '{dateRegexExample}' would match date 2022-07-15</li>
          <li>Date format would thus be '2006-1-2'</li>
        </ol>
      </p>
      In this case the user must set the 'filter' as a valid regular expression
      to match the date. E.g.
      <Typography variant="h6" color="textPrimary">
        Tips
      </Typography>
      <ul>
        <li>
          Try to create filters that are as strict as possible. E.g. matching
          content with 'Google' probably matches many more documents than
          intended. Specific email address or bank account might limit the
          results down.
        </li>
      </ul>
    </HelpButton>
  );
};
