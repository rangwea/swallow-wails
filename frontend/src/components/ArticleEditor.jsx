import React, { useEffect, useState } from "react";
import { message, Drawer, Space, Form, Input, Row, Col, Button } from "antd";
import {
  ArrowLeftOutlined,
  CheckOutlined,
  SettingOutlined,
  EyeOutlined,
  EditOutlined,
  FileImageOutlined,
} from "@ant-design/icons";
import MDEditor from "@uiw/react-md-editor";
import TagInput from "./TagInput";
import { Link, useSearchParams } from "react-router-dom";
import {
  ArticleSave,
  ArticleGet,
  ArticleInsertImage,
  ArticleInsertImageBlob,
} from "../../wailsjs/go/backend/App";
import { getCurrentTime } from "./util";

function ArticleEditor() {
  const [params] = useSearchParams();
  const [id, setId] = useState(params.get("id"));

  // article vars
  const [title, setTitle] = useState(); // title
  const [content, setContent] = useState(""); // content

  // config vars
  const [drawerOpen, setDrawerOpen] = useState(false); // open config drawer?
  const [form] = Form.useForm(); // Form
  
  // editor height
  const [eheight, setEheight] = useState(window.innerHeight - 100);

  // preview button vars
  const [preview, setPreview] = useState("edit");
  const [previewIcon, setPreviewIcon] = useState(<EyeOutlined />);

  const mdTextAreaId = "mdTextArea";

  useEffect(() => {
    window.addEventListener("resize", windowResize);

    init();

    return () => window.removeEventListener("resize", windowResize);
  }, []);

  function windowResize() {
    let editorHeight = window.innerHeight - 100;
    setEheight(editorHeight);
  }

  function init() {
    if (id) {
      // exiestd id，edit
      ArticleGet(id).then((result) => {
        if (result.code != 1) {
          message.error("获取文章失败:" + result.msg);
          return;
        }
        let meta = result.data.meta;
        setTitle(meta.title);
        setContent(result.data.content);
        form.setFieldsValue(meta);
      });
    } else {
      let curDate = getCurrentTime();
      form.setFieldsValue({
        tags: [],
        date: curDate,
        lastmod: curDate,
      });
    }
  }

  function save(e) {
    let meta = getMeta();
    ArticleSave(id, meta, content).then((r) => {
      if (r.code == 1) {
        // success
        setId(r.data);
        message.info("save success");
      } else {
        // fail
        message.error("save fail:", r.msg);
      }
    });
  }

  function showDrawer() {
    setDrawerOpen(true);
  }

  function getMeta() {
    let meta = form.getFieldsValue();
    meta["title"] = title;
    return meta;
  }

  function insertImage() {
    ArticleInsertImage(id).then((r) => {
      insertImageTextToArea(r);
    });
  }

  function insertImageTextToArea(r) {
    if (r.code == 1) {
      const md = insertToTextArea(`![](${r.data})\n`);
      setContent(md);
    }
  }

  function ToolBtns() {
    return (
      <>
        <div style={{ position: "fixed", right: 15, top: 30 }}>
          <Space direction="vertical">
            <Button
              icon={<CheckOutlined />}
              onClick={save}
              shape="circle"
              size="small"
            ></Button>
            {id != "about" && (
              <Button
                icon={<SettingOutlined />}
                onClick={showDrawer}
                shape="circle"
                size="small"
              ></Button>
            )}
            <Button
              icon={<FileImageOutlined />}
              onClick={insertImage}
              shape="circle"
              size="small"
            ></Button>
            <Button
              icon={previewIcon}
              onClick={previewToggle}
              shape="circle"
              size="small"
            ></Button>
            <Link to="/articleList">
              <Button
                icon={<ArrowLeftOutlined />}
                shape="circle"
                size="small"
              ></Button>
            </Link>
          </Space>
        </div>
      </>
    );
  }

  function previewToggle() {
    if (preview == "edit") {
      setPreview("preview");
      setPreviewIcon(<EditOutlined />);
    } else {
      setPreview("edit");
      setPreviewIcon(<EyeOutlined />);
    }
  }

  function insertToTextArea(insertString) {
    const textarea = document.getElementById(mdTextAreaId);
    if (!textarea) {
      return null;
    }

    let sentence = textarea.value;
    const len = sentence.length;
    const pos = textarea.selectionStart;
    const end = textarea.selectionEnd;

    const front = sentence.slice(0, pos);
    const back = sentence.slice(pos, len);

    sentence = front + insertString + back;

    textarea.value = sentence;
    textarea.selectionEnd = end + insertString.length;

    return sentence;
  }

  function onImagePasted(dataTransfer) {
    const file = dataTransfer.files.item(0);
    Promise.resolve(file.arrayBuffer()).then((ab) => {
      const u = new Uint8Array(ab);
      ArticleInsertImageBlob(id, `[${u.toString()}]`).then((r) => {
        insertImageTextToArea(r);
      });
    });
  }

  return (
    <>
      <Row justify="center" style={{ "--wails-draggable": "drag" }}>
        <Col span={16}>
          <Input
            placeholder="Title"
            style={{
              marginTop: 30,
              marginBottom: 5,
              border: "none",
              "--wails-draggable": "no-drag",
            }}
            size="large"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          ></Input>
          <MDEditor
            value={content}
            onChange={setContent}
            onPaste={(e) => {
              if (e.clipboardData.files.length > 0) {
                e.preventDefault();
                onImagePasted(e.clipboardData);
              }
            }}
            onDrop={(e) => {
              e.preventDefault();
              onImagePasted(e.dataTransfer);
            }}
            style={{
              zIndex: 1,
              marginBottom: 10,
              border: "none",
              "--wails-draggable": "no-drag",
            }}
            hideToolbar={true}
            height={eheight}
            preview={preview}
            textareaProps={{
              id: mdTextAreaId,
              placeholder: "Write you text",
            }}
          />
        </Col>
      </Row>
      <ToolBtns></ToolBtns>
      <Drawer
        title="Article Setting"
        placement="right"
        closable={true}
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
        key="drawer"
      >
        <Form form={form} labelCol={{ span: 5 }}>
          <Form.Item label="Tags" name="tags">
            <TagInput></TagInput>
          </Form.Item>
          <Form.Item label="Date" name="date">
            <Input placeholder="Create time"></Input>
          </Form.Item>
          <Form.Item label="Lastmod" name="lastmod">
            <Input placeholder="Last modify time"></Input>
          </Form.Item>
        </Form>
      </Drawer>
    </>
  );
}

export default ArticleEditor;
