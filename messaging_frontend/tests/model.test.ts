import { beforeEach, describe, expect, jest, test } from "@jest/globals";
import { Model } from "../src/model";

//For the functions: updateReaction, addReactions, checkReactions
describe("Reaction model tests", () => {
  let testModel: Model;
  beforeEach(() => {
    testModel = new Model();
    process.env = {
      DATABASE_HOST: "host",
      DATABASE_PATH: "path",
      AUTH_PATH: "auth",
    };
    testModel.token = "token";
  });

  //A mock fetch for these functions
  global.fetch = jest.fn(
    (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
      //Log input
      console.log("Fetching input:", input.toString());
      //Log init
      if (init == null || init.method == null || init.body == null) {
        console.log("No init value");
      } else {
        console.log("init method:", init.method);
        console.log("init body:", init.body);
        console.log("init op:", JSON.parse(init.body.toString())[0].op);
      }
      console.log("Path:", input.toString());
      switch (input.toString()) {
        case `${process.env.DATABASE_HOST}${
          process.env.DATABASE_PATH
        }${"validPath"}`:
          if (init?.method !== "PATCH") {
            console.log("Somethings off...");
          }
          console.log("Valid fetch");
          //Disinguish between an addReactions and updateReactions
          if (init?.body != undefined) {
            if (JSON.parse(init?.body.toString())[0].op == "ObjectAdd") {
              let newuri = "";
              if (init?.body != undefined) {
                newuri = JSON.parse(init?.body.toString()).value;
              }
              return Promise.resolve({
                ok: true,
                json: () =>
                  Promise.resolve({
                    uri: newuri,
                    patchFailed: false,
                    messages: "Add successful",
                  }),
              } as Response);
            } else if (JSON.parse(init?.body.toString())[0].op == "validOp") {
              let newuri = init?.body.toString();
              return Promise.resolve({
                ok: true,
                json: () =>
                  Promise.resolve({
                    uri: newuri,
                    patchFailed: false,
                    messages: "Update successful",
                  }),
              } as Response);
            }
          }
      }
      return Promise.resolve({
        ok: false,
        statusText: "not valid",
        status: 404,
        json: () =>
          Promise.resolve({
            uri: null,
            patchFailed: false,
            messages: "No success",
          }),
      } as Response);
    },
  ) as jest.MockedFunction<typeof fetch>;

  /*
   * Tests updateReaction when the input is valid
   */
  test("updateReaction valid test", async () => {
    let op = "validOp";
    let reaction = "smile";
    let user = "user";
    let path = "validPath";
    const options = {
      method: "PATCH",
      headers: {
        accept: "application/json",
        Authorization: `Bearer token`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify([
        { op: op, path: `/reactions/${reaction}`, value: user },
      ]),
    };
    global.fetch(
      `${process.env.DATABASE_HOST}${process.env.DATABASE_PATH}${path}`,
      options,
    );
    const output = (await testModel.updateReaction(
      op,
      reaction,
      user,
      path,
    )) as any;
    console.log("output is:", output);
    expect(fetch).toHaveBeenCalled();
    expect(output.messages).toBe("Update successful");
  });

  /*
   * Tests updateReaction when the input is invalid
   */
  test("updateReaction invalid test", async () => {
    let op = "invalidOp";
    let reaction = "smile";
    let user = "user";
    let path = "validPath";

    let testError = new Error();
    testModel.updateReaction(op, reaction, user, path).catch((error) => {
      testError = error;
    });
    console.log("Error:", testError.name);
    console.log(testError.message);
    expect(testError.name).toBe("Error");
  });

  test("addReactions valid test", async () => {
    let op = "ObjectAdd";
    let path = "validPath";
    const options = {
      method: "PATCH",
      headers: {
        accept: "application/json",
        Authorization: `Bearer token`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify([
        {
          op: op,
          path: `/reactions`,
          value: {
            smile: [],
            frown: [],
            like: [],
            celebrate: [],
          },
        },
      ]),
    };
    global.fetch(
      `${process.env.DATABASE_HOST}${process.env.DATABASE_PATH}${path}`,
      options,
    );
    const output = (await testModel.addReactions(path)) as any;
    console.log("output is:", output);
    expect(fetch).toHaveBeenCalled();
    expect(output.messages).toBe("Add successful");
  });

  /*
   * Tests addReactions when the input is invalid
   */
  test("addReactions invalid test", async () => {
    let path = "validPath";
    let testError = new Error();
    testModel.addReactions(path).catch((error) => {
      testError = error;
    });
    console.log("Error:", testError.name);
    console.log(testError.message);
    expect(testError.name).toBe("Error");
  });

  /*
   * Tests checkReactions when the input is empty
   */
  test("checkReactions empty test", async () => {
    let response = { doc: "", path: "validPath" };

    let op = "ObjectAdd";
    let path = "validPath";
    const options = {
      method: "PATCH",
      headers: {
        accept: "application/json",
        Authorization: `Bearer token`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify([
        {
          op: op,
          path: `/reactions`,
          value: {
            smile: [],
            frown: [],
            like: [],
            celebrate: [],
          },
        },
      ]),
    };
    global.fetch(
      `${process.env.DATABASE_HOST}${process.env.DATABASE_PATH}${path}`,
      options,
    );
    await testModel.checkReactions(response);
    expect(fetch).toHaveBeenCalled();
  });

  /*
   * Tests checkReactions when the input is already filled
   */
  test("checkReactions with reaction field", async () => {
    let response = { doc: { reactions: [] }, path: "newPath" };
    let op = "ObjectAdd";
    let path = "newPath";
    const options = {
      method: "PATCH",
      headers: {
        accept: "application/json",
        Authorization: `Bearer token`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify([
        {
          op: op,
          path: `/reactions`,
          value: {
            smile: [],
            frown: [],
            like: [],
            celebrate: [],
          },
        },
      ]),
    };
    console.log("Test 6");
    await testModel.checkReactions(response);
    expect(fetch).not.toHaveBeenCalledWith(
      `${process.env.DATABASE_HOST}${process.env.DATABASE_PATH}${path}`,
      options,
    );
  });
});
