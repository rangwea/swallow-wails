import React, { useState, useRef } from "react";
import { Tag, Input } from "antd";

const TagInput = (props) => {
  const [value, setValue] = useState(props.value || []);
  const [valueInput, setValueInput] = useState("");
  const inputRef = useRef(null);

  function pressEnter(e) {
    if (e.target.value) {
      let new_value = [...value, e.target.value]
      setValue(new_value);
      setValueInput("");
      props.onChange(new_value);
    }
  }

  function preventDefault(str, e) {
    e.preventDefault();
    setValue(value.filter((item) => item !== str));
  }

  function focus() {
    inputRef.current && inputRef.current.focus();
  }

  function handleChange(e) {
    let elm = e.target;
    setValueInput(elm.value);
  }

  function keyDown(e) {
    if (e.keyCode === 8 && !valueInput) {
      setValue(
        value.filter(function (v, i, ar) {
          return i !== ar.length - 1;
        })
      );
    }
  }

  return (
    <div>
      <div onClick={focus} className="tagInputWrap">
        <ul className="tagInputUlClass">
          {value &&
            value.map((item, index) => (
              <li key={index} style={{ float: "left", marginBottom: "8px" }}>
                <Tag closable onClose={(e) => preventDefault(item, e)}>
                  {item}
                </Tag>
              </li>
            ))}
          <li style={{ float: "left" }}>
            <Input
              onKeyDown={keyDown}
              ref={inputRef}
              value={valueInput}
              className="tagInput"
              onPressEnter={pressEnter}
              onChange={handleChange}
            />
          </li>
        </ul>
      </div>
    </div>
  );
};

export default TagInput;