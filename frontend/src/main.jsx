import React from "react";
import ReactDOM from "react-dom/client";
import { createHashRouter, RouterProvider } from "react-router-dom";
import "./style.css";
import ArticleList from "./components/ArticleList";
import ArticleEditor from "./components/ArticleEditor";
import Config from "./components/Config";

const router = createHashRouter([
  {
    path: "/",
    element: <ArticleList />,
  },
  {
    path: "/articleList",
    element: <ArticleList />,
  },
  {
    path: "/articleEditor",
    element: <ArticleEditor />,
  },
  {
    path: "/config",
    element: <Config />,
  },
]);

ReactDOM.createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <RouterProvider router={router}></RouterProvider>
  </React.StrictMode>
);
