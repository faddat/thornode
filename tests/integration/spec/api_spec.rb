require_relative './helper.rb'

describe "API Tests" do

  context "Check /ping responds" do
    it "should return 'pong'" do
      resp = get("/ping")
      expect(resp.body['ping']).to eq "pong"
    end
  end

  context "Check that an empty tx hash returns properly" do
    it "should have no values" do
      resp = get("/tx/bogus")
      expect(resp.body['request']).to eq ""
      expect(resp.body['status']).to eq ""
      expect(resp.body['txhash']).to eq ""
    end
  end

  context "Create a pool" do
    it "create a pool for bnb" do
      resp = processTx("AF64E866F7EDD74A558BF1519FB12700DDE51CD0DB5166ED37C568BE04E0C7F3")
      expect(resp.code).to eq("200"), "Are you working from a clean blockchain? \n(#{resp.code}: #{resp.body})"
    end

    it "should get a conflict the second time" do
      resp = processTx("AF64E866F7EDD74A558BF1519FB12700DDE51CD0DB5166ED37C568BE04E0C7F3")
      expect(resp.code).to eq("500")
    end

    it "should be able to get the pool" do
      resp = get("/pool/TCAN-014")
      expect(resp.body['pool_id']).to eq("pool-TCAN-014"), resp.body
    end

    it "should show up in listing of pools" do
      resp = get("/pools")
      expect(resp.body[1]['pool_id']).to eq("pool-TCAN-014"), resp.body
    end

  end

end
