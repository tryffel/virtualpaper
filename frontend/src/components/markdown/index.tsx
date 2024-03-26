import React, { Suspense } from "react";
import { Loading } from "react-admin";
const Editor = React.lazy(() => import("./Editor.tsx"));
const Field = React.lazy(() => import("./Field.tsx"));

export type MarkdownProps = {
  source: string;
  label?: string;
};

export const MarkdownInput = (props: MarkdownProps) => {
  return (
    <Suspense fallback={<Loading />}>
      <Editor {...props} />
    </Suspense>
  );
};

export const MarkdownField = (props: MarkdownProps) => {
  return (
    <Suspense fallback={<Loading />}>
      <Field {...props} />
    </Suspense>
  );
};
