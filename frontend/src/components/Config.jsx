import React, { useEffect, useState } from "react";
import {
  message,
  Button,
  Form,
  Space,
  Tabs,
  Input,
  Select,
  Row,
  Col,
  Layout,
} from "antd";
import { Link } from "react-router-dom";
import {
  SiteConfigGet,
  SiteConfigSave,
  ConfSave,
  ConfGet,
  SelectConfImage,
  ConfGetThemes,
} from "../../wailsjs/go/backend/App";

const { Header, Footer, Content } = Layout;

function Config() {
  const [websiteForm] = Form.useForm();
  const [githubForm] = Form.useForm();
  const [currentTabKey, setCurrentTabKey] = useState("website");
  const [avatar, setAvatar] = useState("/static/images/avatar.png");
  const [favicon, setFavicon] = useState("/static/images/favicon.ico");
  const [themeOptions, setThemeOptions] = useState([]);

  useEffect(() => {
    getThemes();
    getConfigData(currentTabKey);
  }, []);

  const onChange = (activeKey) => {
    setCurrentTabKey(activeKey);
    getConfigData(activeKey);
  };

  const getThemes = () => {
    ConfGetThemes().then((result) => {
      if (result.code === 0) {
        message.error("get themes fail:" + result.msg);
      } else {
        let ts = [];
        for (const x of result.data) {
          ts.push({
            value: x,
            label: x,
          });
        }
        setThemeOptions(ts);
      }
    });
  };

  const getConfigData = (type) => {
    if (type === "website") {
      SiteConfigGet().then((result) => {
        if (result.code === 0) {
          message.error("get website config fail:" + result.msg);
        } else {
          websiteForm.setFieldsValue(result.data);
        }
      });
    } else if (type === "github") {
      ConfGet("github").then((r) => {
        if (r.code === 0) {
          message.error("get config fail:" + result.msg);
        } else {
          githubForm.setFieldsValue(r.data);
        }
      });
    }
  };

  const saveConfigData = () => {
    if (currentTabKey === "website") {
      let c = websiteForm.getFieldsValue();
      SiteConfigSave(c).then((r) => {
        if (r.code === 0) {
          message.error(r.msg);
        } else {
          message.info("save success", 1);
        }
      });
    } else if (currentTabKey === "github") {
      let c = githubForm.getFieldsValue();
      ConfSave("github", c).then((r) => {
        if (r.code === 0) {
          message.error(r.msg);
        } else {
          message.info("save success", 1);
        }
      });
    }
  };

  function setImage(src, setMethod) {
    SelectConfImage(src).then((r) => {
      if (r.code === 1) {
        setMethod(`${src}?${new Date()}`);
      } else {
        message.error(r.msg);
      }
    });
  }

  const siteTab = () => {
    return (
      <>
        <Row justify="center">
          <Col span={18}>
            <Form labelCol={{ span: 3 }} form={websiteForm}>
              <Form.Item label="Title" name="title">
                <Input />
              </Form.Item>
              <Form.Item label="Description" name="description">
                <Input />
              </Form.Item>
              <Form.Item label="Theme" name="theme">
                <Select defaultValue="stack" options={themeOptions} />
              </Form.Item>
              <Form.Item label="Language" name="defaultContentLanguage">
                <Select
                  defaultValue="en"
                  options={[
                    {
                      value: "en",
                      label: "en",
                    },
                    {
                      value: "zh",
                      label: "中文",
                    },
                  ]}
                />
              </Form.Item>
              <Form.Item label="Copyright" name="copyright">
                <Input />
              </Form.Item>
              <Form.Item label="Author" name={["params", "author", "name"]}>
                <Input />
              </Form.Item>
              <Form.Item label="About Me" name="about">
                <Link to={"/articleEditor?id=about"}> Go To Edit </Link>
              </Form.Item>
            </Form>
          </Col>
        </Row>
        <Row justify="center" gutter={20}>
          <Col>
            <div style={{ marginBottom: 5 }}>select avatar:</div>
            <img
              className="confImg"
              src={avatar}
              onClick={(e) => setImage("/static/images/avatar.png", setAvatar)}
            />
          </Col>
          <Col>
            <div style={{ marginBottom: 5 }}>select favicon:</div>
            <img
              className="confImg"
              src={favicon}
              onClick={(e) =>
                setImage("/static/images/favicon.ico", setFavicon)
              }
            />
          </Col>
        </Row>
        <Row>
          <Col>
            {/* todo: edit hugo.toml file by toml editor. */}
            <Button>Advanced Configuration</Button>
          </Col>
        </Row>
      </>
    );
  };

  const githubTab = () => {
    return (
      <Row justify="center">
        <Col span={18}>
          <Form labelCol={{ span: 4 }} form={githubForm}>
            <Form.Item label="Repository" name="repository">
              <Input />
            </Form.Item>
            <Form.Item label="Email" name="email">
              <Input />
            </Form.Item>
            <Form.Item label="Username" name="username">
              <Input />
            </Form.Item>
            <Form.Item label="Token" name="token">
              <Input />
            </Form.Item>
            <Form.Item label="CNAME" name="cname">
              <Input />
            </Form.Item>
            <Form.Item>
              <Button>Connection Test</Button>
            </Form.Item>
          </Form>
        </Col>
      </Row>
    );
  };

  const items = [
    {
      key: "website",
      label: "Website",
      children: siteTab(),
    },
    {
      key: "github",
      label: "Github",
      children: githubTab(),
    },
  ];

  return (
    <Layout style={{ height: "100vh", background: "#FFF" }}>
      <Header
        style={{
          height: 50,
          "--wails-draggable": "drag",
          background: "#FFF",
        }}
      ></Header>
      <Content>
        <Tabs
          defaultActiveKey="1"
          items={items}
          tabPosition="left"
          style={{ height: "100%" }}
          onChange={onChange}
        />
      </Content>
      <Footer
        style={{
          background: "#FFF",
          height: 50,
          marginBottom: 20,
        }}
      >
        <Row justify="center">
          <Col>
            <Space>
              <Link to="/articleList">
                <Button>Cancle</Button>
              </Link>
              <Button onClick={saveConfigData} type="primary">
                Save
              </Button>
            </Space>
          </Col>
        </Row>
      </Footer>
    </Layout>
  );
}

export default Config;
