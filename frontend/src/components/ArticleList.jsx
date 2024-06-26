import React, { useState, useEffect } from "react";
import {
  FloatButton,
  Row,
  Input,
  List,
  Space,
  Col,
  message,
  Button,
  Checkbox,
} from "antd";
import {
  PlusOutlined,
  SearchOutlined,
  FieldTimeOutlined,
  TagsOutlined,
  SettingOutlined,
  CloudUploadOutlined,
  EyeOutlined,
  DeleteOutlined,
} from "@ant-design/icons";
import { Link, useNavigate } from "react-router-dom";
import {
  ArticleList,
  ArticleRemove,
  SitePreview,
  SiteDeploy,
} from "../../wailsjs/go/backend/App";

const { Search } = Input;
const IconText = ({ icon, text }) => (
  <Space>
    {React.createElement(icon)}
    {text}
  </Space>
);

function ArtistList() {
  const navigate = useNavigate();
  const [articles, setArticles] = useState([]);
  const [checked, setChecked] = useState([]);
  const [deleteBtnShow, setDeleteBtnShow] = useState(false);

  function searchArticles(v, e) {
    ArticleList(v).then((result) => {
      setArticles(result.data);
    });
  }

  useEffect(() => {
    searchArticles("", null);
  }, []);

  function preview() {
    SitePreview().then((r) => {
      if (r.code !== 1) {
        message.error(r.msg);
      }
    });
  }

  function deploy() {
    SiteDeploy().then((r) => {
      if (r.code !== 1) {
        message.error(r.msg);
      }
    });
  }

  function ToolBtns() {
    return (
      <>
        <div style={{ position: "fixed", left: 15, bottom: 15 }}>
          <Space direction="vertical">
            <Link to="/config">
              <Button
                icon={<SettingOutlined />}
                shape="circle"
                size="small"
              ></Button>
            </Link>
            <Button
              icon={<EyeOutlined />}
              onClick={preview}
              shape="circle"
              size="small"
            ></Button>
            <Button
              icon={<CloudUploadOutlined />}
              onClick={deploy}
              shape="circle"
              size="small"
            ></Button>
          </Space>
        </div>
      </>
    );
  }

  function removeArticle() {
    if (checked.length > 0) {
      ArticleRemove(checked).then((r) => {
        if (r.code === 1) {
          message.info(`removed ${checked.length} articles`, 2);
          searchArticles("", null);
          setChecked([]);
          setDeleteBtnShow(false);
        } else {
          message.error(r.msg);
        }
      });
    }
  }

  return (
    <>
      <Row
        justify="center"
        style={{ paddingTop: 20, "--wails-draggable": "drag" }}
      >
        <Col span={4} style={{ paddingTop: 8 }}>
          {deleteBtnShow ? (
            <Button
              icon={<DeleteOutlined />}
              size="small"
              onClick={removeArticle}
            >
              selected {checked.length}
            </Button>
          ) : null}
        </Col>
        <Col span={8}>
          <Search
            placeholder="search"
            prefix={<SearchOutlined />}
            onSearch={searchArticles}
            style={{ "--wails-draggable": "no-drag" }}
          />
        </Col>
        <Col span={4}></Col>
      </Row>
      <Row justify="center">
        <Col span={16}>
          <Checkbox.Group
            style={{ width: "100%" }}
            value={checked}
            onChange={(checkedValues) => {
              setDeleteBtnShow(checkedValues.length !== 0);
              setChecked(checkedValues);
            }}
          >
            <List
              size="large"
              style={{ width: "100%" }}
              header={<div></div>}
              pagination={{
                pageSize: 10,
              }}
              dataSource={articles}
              renderItem={(item) => (
                <List.Item
                  key={item.id}
                  actions={[
                    <IconText
                      icon={FieldTimeOutlined}
                      text={item.createTime}
                      key={item.id + "-createTime"}
                    />,
                    <IconText
                      icon={TagsOutlined}
                      text={item.tags}
                      key={item.id + "-tags"}
                    />,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<Checkbox value={item.id + ""} />}
                    title={
                      <Link to={"/articleEditor?id=" + item.id}>
                        {item.title}
                      </Link>
                    }
                    key={item.id}
                    description={item.description}
                  />
                </List.Item>
              )}
            />
          </Checkbox.Group>
        </Col>
      </Row>
      <Link to="/articleEditor">
        <FloatButton
          type="primary"
          icon={<PlusOutlined />}
          style={{
            right: 30,
            bottom: 30,
          }}
        ></FloatButton>
      </Link>
      <ToolBtns></ToolBtns>
    </>
  );
}

export default ArtistList;
