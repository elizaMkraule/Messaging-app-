import Ajv from "../node_modules/ajv/dist/ajv";
import postSchema from "../schemas/post.json";
import channelSchema from "../schemas/channel.json";
import workspaceSchema from "../schemas/workspace.json";

/**
 * Class Validator validates the incoming data from the database
 */
export class Validator {
  private validateChannel;
  private validateWorkspace;
  private validatePost;
  private ajv: Ajv;

  constructor() {
    const Ajv = require("ajv");
    this.ajv = new Ajv();
    this.validateChannel = this.ajv.compile(channelSchema);
    this.validateWorkspace = this.ajv.compile(workspaceSchema);
    this.validatePost = this.ajv.compile(postSchema);
  }

  /**
   *  validate validates the json data with a schema depending on the type passed in
   *
   * @param data the json data to be validates
   * @param type the type of schema to be used
   *
   *
   * @return a boolan indicating if the data conforms to a schema
   */
  public validate(data: any, type: "channel" | "workspace" | "post"): boolean {
    let isValid: boolean;
    switch (type) {
      case "channel":
        isValid = this.validateChannel(data) as boolean;
        break;
      case "workspace":
        isValid = this.validateWorkspace(data) as boolean;
        break;
      case "post":
        isValid = this.validatePost(data);
        break;
      default:
        throw new Error(`Unknown type for validation: ${type}`);
    }

    if (!isValid) {
      console.error("Validation error:", this.ajv.errorsText(this.ajv.errors));
      throw new Error("Invalid data:");
    }

    return isValid;
  }
}
